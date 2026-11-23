package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// OIDCConfig は OIDC のプロバイダー（Google など）とのやり取りに使う設定。
type OIDCConfig struct {
	Issuer       string // 例：https://accounts.google.com
	ClientID     string
	ClientSecret string
	RedirectURL  string // 例：https://relay.example.com/api/auth/callback
}

// OIDC は認可コードフロー（PKCE）でログインし、ID トークンを検証する。
type OIDC struct {
	cfg      OIDCConfig
	authURL  string
	tokenURL string
	issuers  []string
	keyfunc  jwt.Keyfunc
	client   *http.Client
}

type discovery struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
}

// NewOIDC は Issuer の /.well-known/openid-configuration を読み、公開鍵（JWKS）の取得を始める。
func NewOIDC(ctx context.Context, cfg OIDCConfig) (*OIDC, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	var d discovery
	if err := getJSON(ctx, client, strings.TrimSuffix(cfg.Issuer, "/")+"/.well-known/openid-configuration", &d); err != nil {
		return nil, fmt.Errorf("oidc discovery: %w", err)
	}
	if d.AuthorizationEndpoint == "" || d.TokenEndpoint == "" || d.JWKSURI == "" {
		return nil, errors.New("oidc discovery: missing endpoints")
	}
	k, err := keyfunc.NewDefaultCtx(ctx, []string{d.JWKSURI})
	if err != nil {
		return nil, fmt.Errorf("load JWKS: %w", err)
	}
	return &OIDC{
		cfg:      cfg,
		authURL:  d.AuthorizationEndpoint,
		tokenURL: d.TokenEndpoint,
		issuers:  acceptedIssuers(d.Issuer),
		keyfunc:  k.Keyfunc,
		client:   client,
	}, nil
}

// acceptedIssuers は ID トークンの iss として受け入れる値。Google は "accounts.google.com" も使う。
func acceptedIssuers(issuer string) []string {
	out := []string{issuer}
	if issuer == "https://accounts.google.com" {
		out = append(out, "accounts.google.com")
	}
	return out
}

// AuthCodeURL はプロバイダーのログイン画面の URL を返す。
func (o *OIDC) AuthCodeURL(state, nonce, verifier string) string {
	q := url.Values{
		"response_type":         {"code"},
		"client_id":             {o.cfg.ClientID},
		"redirect_uri":          {o.cfg.RedirectURL},
		"scope":                 {"openid email profile"},
		"state":                 {state},
		"nonce":                 {nonce},
		"code_challenge":        {pkceChallenge(verifier)},
		"code_challenge_method": {"S256"},
		// アカウントを選ぶ画面を出す（複数の Google アカウントを使い分けられるように）
		"prompt": {"select_account"},
	}
	return o.authURL + "?" + q.Encode()
}

// Exchange は認可コードを ID トークンに替える。アクセストークンとリフレッシュトークンは使わない。
func (o *OIDC) Exchange(ctx context.Context, code, verifier string) (string, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {o.cfg.RedirectURL},
		"client_id":     {o.cfg.ClientID},
		"client_secret": {o.cfg.ClientSecret},
		"code_verifier": {verifier},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	res, err := o.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return "", fmt.Errorf("token endpoint: %s: %s", res.Status, body)
	}
	var t struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
		return "", err
	}
	if t.IDToken == "" {
		return "", errors.New("token endpoint: id_token is missing")
	}
	return t.IDToken, nil
}

type idClaims struct {
	Email         string          `json:"email"`
	EmailVerified json.RawMessage `json:"email_verified"`
	Name          string          `json:"name"`
	Nonce         string          `json:"nonce"`
	jwt.RegisteredClaims
}

// VerifyIDToken は ID トークンの署名・iss・aud・exp・nonce と、メールアドレスが確認済みであることを検証する。
func (o *OIDC) VerifyIDToken(raw, nonce string) (Principal, error) {
	var c idClaims
	_, err := jwt.ParseWithClaims(raw, &c, o.keyfunc,
		jwt.WithValidMethods([]string{"RS256", "ES256"}),
		jwt.WithAudience(o.cfg.ClientID),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		return Principal{}, err
	}
	if !slices.Contains(o.issuers, c.Issuer) {
		return Principal{}, fmt.Errorf("unexpected issuer %q", c.Issuer)
	}
	if c.Nonce == "" || c.Nonce != nonce {
		return Principal{}, errors.New("nonce mismatch")
	}
	if c.Subject == "" {
		return Principal{}, errors.New("sub claim is missing")
	}
	// Google は真偽値、プロバイダーによっては文字列で返す
	if v := strings.Trim(string(c.EmailVerified), `"`); c.Email == "" || v != "true" {
		return Principal{}, errors.New("email is not verified")
	}
	// 同じ人は、iss の書き方が違っても同じ Issuer として扱う
	return Principal{Issuer: o.issuers[0], Subject: c.Subject, Email: c.Email, Name: c.Name}, nil
}

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func getJSON(ctx context.Context, client *http.Client, u string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", u, res.Status)
	}
	return json.NewDecoder(res.Body).Decode(v)
}
