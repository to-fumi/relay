package auth_test

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"relay/internal/platform/auth"

	"github.com/google/uuid"
)

type memSessions struct {
	byID map[string]struct {
		user uuid.UUID
		exp  time.Time
	}
}

func (m *memSessions) Create(_ context.Context, user uuid.UUID, now time.Time, ttl time.Duration) (string, error) {
	id := uuid.NewString()
	m.byID[id] = struct {
		user uuid.UUID
		exp  time.Time
	}{user, now.Add(ttl)}
	return id, nil
}

func (m *memSessions) Lookup(_ context.Context, id string, now time.Time) (uuid.UUID, error) {
	s, ok := m.byID[id]
	if !ok || !now.Before(s.exp) {
		return uuid.Nil, auth.ErrNoSession
	}
	return s.user, nil
}

// memUsers は Issuer と Subject の組ごとにユーザー ID を払い出す。
type memUsers map[string]uuid.UUID

func (m memUsers) ResolveUser(_ context.Context, p auth.Principal) (uuid.UUID, error) {
	key := p.Issuer + "|" + p.Subject
	if id, ok := m[key]; ok {
		return id, nil
	}
	m[key] = uuid.New()
	return m[key], nil
}

type env struct {
	srv      *httptest.Server
	client   *http.Client
	provider *fakeProvider
	users    memUsers
	now      time.Time
}

// newEnv はログインの /api/auth/* と、セッションが要る /api/me（ユーザー ID を返すだけ）を持つサーバーを立てる。
func newEnv(t *testing.T, dev bool, allowed ...string) *env {
	t.Helper()
	e := &env{users: memUsers{}, now: time.Now()}
	sessions := &memSessions{byID: map[string]struct {
		user uuid.UUID
		exp  time.Time
	}{}}
	cfg := auth.HandlersConfig{
		Dev: dev, Store: sessions, Users: e.users, Allowed: auth.NewAllowlist(allowed),
		TTL: 24 * time.Hour, Now: func() time.Time { return e.now },
	}
	if !dev {
		e.provider = newFakeProvider(t)
		cfg.Provider = e.provider.oidc(t)
	}
	mux := http.NewServeMux()
	auth.NewHandlers(cfg).Register(mux)
	mux.Handle("/api/me", auth.Middleware(sessions, func() time.Time { return e.now })(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			id, _ := auth.UserIDFrom(r.Context())
			_, _ = w.Write([]byte(id.String()))
		})))
	e.srv = httptest.NewServer(mux)
	t.Cleanup(e.srv.Close)
	jar, _ := cookiejar.New(nil)
	e.client = &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return e
}

// result は応答の状態とヘッダー（本文は使わないので、受け取ったらすぐ閉じる）。
type result struct {
	StatusCode int
	Header     http.Header
	cookies    []*http.Cookie
}

func (r result) Cookies() []*http.Cookie { return r.cookies }

func (e *env) do(t *testing.T, req *http.Request) result {
	t.Helper()
	res, err := e.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	return result{StatusCode: res.StatusCode, Header: res.Header, cookies: res.Cookies()}
}

func (e *env) get(t *testing.T, path string) result {
	t.Helper()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, e.srv.URL+path, http.NoBody)
	return e.do(t, req)
}

func (e *env) post(t *testing.T, path, body string) result {
	t.Helper()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, e.srv.URL+path, strings.NewReader(body))
	return e.do(t, req)
}

// loginVia は /api/auth/login → プロバイダー → /api/auth/callback の流れをたどり、callback の応答を返す。
// プロバイダーがログインの後に返す ID トークンの内容は change で変えられる。
func (e *env) loginVia(t *testing.T, change func(nonce string) map[string]any) result {
	t.Helper()
	res := e.get(t, "/api/auth/login")
	if res.StatusCode != http.StatusFound {
		t.Fatalf("login status = %d", res.StatusCode)
	}
	authURL, _ := url.Parse(res.Header.Get("Location"))
	q := authURL.Query()
	claims := e.provider.validClaims(q.Get("nonce"))
	if change != nil {
		for k, v := range change(q.Get("nonce")) {
			claims[k] = v
		}
	}
	e.provider.mu.Lock()
	e.provider.claims["code-1"] = claims
	e.provider.mu.Unlock()
	return e.get(t, "/api/auth/callback?code=code-1&state="+url.QueryEscape(q.Get("state")))
}

func TestOIDCLogin_CreatesSession(t *testing.T) {
	e := newEnv(t, false, "sato@example.com")

	if res := e.get(t, "/api/me"); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("before login: status = %d; want 401", res.StatusCode)
	}
	res := e.loginVia(t, nil)
	if res.StatusCode != http.StatusFound || res.Header.Get("Location") != "/" {
		t.Fatalf("callback = %d %s; want 302 /", res.StatusCode, res.Header.Get("Location"))
	}
	// PKCE：ログインのときの verifier がトークンのエンドポイントに送られている
	if e.provider.lastVerifier == "" {
		t.Fatal("code_verifier was not sent")
	}
	var session *http.Cookie
	for _, c := range res.Cookies() {
		if c.Name == auth.SessionCookie {
			session = c
		}
	}
	if session == nil || !session.HttpOnly || session.SameSite != http.SameSiteLaxMode || session.MaxAge != 24*60*60 {
		t.Fatalf("session cookie = %+v", session)
	}
	if res := e.get(t, "/api/me"); res.StatusCode != http.StatusOK {
		t.Fatalf("after login: status = %d; want 200", res.StatusCode)
	}

	// 期限（1 日）が過ぎると 401
	e.now = e.now.Add(24*time.Hour + time.Second)
	if res := e.get(t, "/api/me"); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("after expiry: status = %d; want 401", res.StatusCode)
	}
}

func TestOIDCLogin_Failures(t *testing.T) {
	tests := []struct {
		name    string
		allowed string
		change  func(nonce string) map[string]any
		want    string
	}{
		{name: "not allowed", allowed: "suzuki@example.com", want: "/?login=forbidden"},
		{name: "allowed by domain", allowed: "@example.com", want: "/"},
		{name: "wrong nonce", allowed: "@example.com", change: func(string) map[string]any { return map[string]any{"nonce": "x"} }, want: "/?login=failed"},
		{name: "email not verified", allowed: "@example.com", change: func(string) map[string]any { return map[string]any{"email_verified": false} }, want: "/?login=failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t, false, tt.allowed)
			res := e.loginVia(t, tt.change)
			if res.StatusCode != http.StatusFound || res.Header.Get("Location") != tt.want {
				t.Fatalf("callback = %d %s; want 302 %s", res.StatusCode, res.Header.Get("Location"), tt.want)
			}
			wantStatus := http.StatusUnauthorized
			if tt.want == "/" {
				wantStatus = http.StatusOK
			}
			if res := e.get(t, "/api/me"); res.StatusCode != wantStatus {
				t.Fatalf("/api/me status = %d; want %d", res.StatusCode, wantStatus)
			}
		})
	}
}

func TestOIDCCallback_RejectsWrongState(t *testing.T) {
	e := newEnv(t, false, "@example.com")
	e.get(t, "/api/auth/login")
	e.provider.claims["code-1"] = e.provider.validClaims("whatever")
	res := e.get(t, "/api/auth/callback?code=code-1&state=forged")
	if res.Header.Get("Location") != "/?login=failed" {
		t.Fatalf("Location = %s; want /?login=failed", res.Header.Get("Location"))
	}
}

func TestDevLogin(t *testing.T) {
	e := newEnv(t, true)
	post := func(body string) result { return e.post(t, "/api/auth/dev-login", body) }
	if res := post(`{"email":"no-at-mark"}`); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid email: status = %d; want 400", res.StatusCode)
	}
	if res := post(`{"email":"Sato@Example.com"}`); res.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d; want 204", res.StatusCode)
	}
	if res := e.get(t, "/api/me"); res.StatusCode != http.StatusOK {
		t.Fatalf("/api/me status = %d; want 200", res.StatusCode)
	}
	// 大文字・小文字が違っても同じ人
	if _, ok := e.users["dev|sato@example.com"]; !ok || len(e.users) != 1 {
		t.Fatalf("users = %v", e.users)
	}
	// OIDC のログインは使えない
	if res := e.get(t, "/api/auth/login"); res.StatusCode != http.StatusNotFound {
		t.Fatalf("oidc login in dev: status = %d; want 404", res.StatusCode)
	}
}

func TestDevLogin_DisabledInOIDCMode(t *testing.T) {
	e := newEnv(t, false, "@example.com")
	if res := e.post(t, "/api/auth/dev-login", `{"email":"a@b.c"}`); res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d; want 404", res.StatusCode)
	}
}

func TestAllowlist(t *testing.T) {
	a := auth.NewAllowlist([]string{" Sato@Example.com ", "@corp.example"})
	for email, want := range map[string]bool{
		"sato@example.com":    true,
		"SATO@EXAMPLE.COM":    true,
		"suzuki@example.com":  false,
		"anyone@corp.example": true,
		"x@evilcorp.example":  false,
	} {
		if got := a.Allows(email); got != want {
			t.Errorf("Allows(%q) = %v; want %v", email, got, want)
		}
	}
	if auth.NewAllowlist(nil).Allows("a@b.c") {
		t.Error("empty allowlist allowed someone")
	}
}
