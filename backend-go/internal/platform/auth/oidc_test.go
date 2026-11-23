package auth_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"relay/internal/platform/auth"

	"github.com/golang-jwt/jwt/v5"
)

// fakeProvider は OIDC のプロバイダー（Google の代わり）。discovery・JWKS・トークンのエンドポイントを持つ。
type fakeProvider struct {
	t      *testing.T
	srv    *httptest.Server
	key    *rsa.PrivateKey
	mu     sync.Mutex
	claims map[string]jwt.MapClaims // code → ID トークンに入れる内容
	// lastVerifier は最後にトークンのエンドポイントへ送られた code_verifier
	lastVerifier string
}

const (
	clientID = "client-id"
	kid      = "test-key"
)

func newFakeProvider(t *testing.T) *fakeProvider {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	p := &fakeProvider{t: t, key: key, claims: map[string]jwt.MapClaims{}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"issuer":                 p.srv.URL,
			"authorization_endpoint": p.srv.URL + "/authorize",
			"token_endpoint":         p.srv.URL + "/token",
			"jwks_uri":               p.srv.URL + "/jwks",
		})
	})
	mux.HandleFunc("GET /jwks", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
			"kty": "RSA", "alg": "RS256", "use": "sig", "kid": kid,
			"n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
			"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
		}}})
	})
	mux.HandleFunc("POST /token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		p.mu.Lock()
		defer p.mu.Unlock()
		p.lastVerifier = r.PostForm.Get("code_verifier")
		c, ok := p.claims[r.PostForm.Get("code")]
		if !ok || r.PostForm.Get("client_secret") != "secret" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"id_token": p.sign(c, key)})
	})
	p.srv = httptest.NewServer(mux)
	t.Cleanup(p.srv.Close)
	return p
}

func (p *fakeProvider) sign(c jwt.MapClaims, key *rsa.PrivateKey) string {
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, c)
	tok.Header["kid"] = kid
	s, err := tok.SignedString(key)
	if err != nil {
		p.t.Fatal(err)
	}
	return s
}

// validClaims は検証を通る ID トークンの内容。
func (p *fakeProvider) validClaims(nonce string) jwt.MapClaims {
	return jwt.MapClaims{
		"iss": p.srv.URL, "aud": clientID, "sub": "1234567890",
		"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(),
		"nonce": nonce, "email": "Sato@Example.com", "email_verified": true, "name": "佐藤",
	}
}

func (p *fakeProvider) oidc(t *testing.T) *auth.OIDC {
	t.Helper()
	o, err := auth.NewOIDC(context.Background(), auth.OIDCConfig{
		Issuer: p.srv.URL, ClientID: clientID, ClientSecret: "secret", RedirectURL: "https://relay.test/api/auth/callback",
	})
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func TestVerifyIDToken(t *testing.T) {
	p := newFakeProvider(t)
	o := p.oidc(t)
	other, _ := rsa.GenerateKey(rand.Reader, 2048)

	tests := []struct {
		name   string
		change func(jwt.MapClaims)
		key    *rsa.PrivateKey
		ok     bool
	}{
		{name: "valid", ok: true},
		{name: "email_verified as string", change: func(c jwt.MapClaims) { c["email_verified"] = "true" }, ok: true},
		{name: "wrong signature", key: other},
		{name: "wrong issuer", change: func(c jwt.MapClaims) { c["iss"] = "https://evil.example" }},
		{name: "wrong audience", change: func(c jwt.MapClaims) { c["aud"] = "other-client" }},
		{name: "expired", change: func(c jwt.MapClaims) { c["exp"] = time.Now().Add(-time.Hour).Unix() }},
		{name: "no exp", change: func(c jwt.MapClaims) { delete(c, "exp") }},
		{name: "wrong nonce", change: func(c jwt.MapClaims) { c["nonce"] = "other" }},
		{name: "email not verified", change: func(c jwt.MapClaims) { c["email_verified"] = false }},
		{name: "no sub", change: func(c jwt.MapClaims) { delete(c, "sub") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := p.validClaims("n1")
			if tt.change != nil {
				tt.change(c)
			}
			key := p.key
			if tt.key != nil {
				key = tt.key
			}
			got, err := o.VerifyIDToken(p.sign(c, key), "n1")
			if tt.ok != (err == nil) {
				t.Fatalf("err = %v; want ok=%v", err, tt.ok)
			}
			if tt.ok && (got.Issuer != p.srv.URL || got.Subject != "1234567890" || got.Email != "Sato@Example.com" || got.Name != "佐藤") {
				t.Fatalf("principal = %+v", got)
			}
		})
	}
}

func TestAuthCodeURL_UsesPKCE(t *testing.T) {
	p := newFakeProvider(t)
	u, err := url.Parse(p.oidc(t).AuthCodeURL("st", "no", "verifier"))
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if !strings.HasSuffix(u.Path, "/authorize") || q.Get("client_id") != clientID || q.Get("state") != "st" ||
		q.Get("nonce") != "no" || q.Get("code_challenge_method") != "S256" || q.Get("code_challenge") == "" ||
		q.Get("code_challenge") == "verifier" || !strings.Contains(q.Get("scope"), "openid") {
		t.Fatalf("unexpected auth URL: %s", u)
	}
}
