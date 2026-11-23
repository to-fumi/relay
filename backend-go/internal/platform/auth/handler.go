package auth

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"relay/internal/platform/httpx"
)

// stateCookie はログインの途中（プロバイダーへ行って戻るまで）の state・nonce・PKCE を入れる Cookie。
const stateCookie = "relay_oidc"

// Provider は OIDC のプロバイダーとのやり取り（*OIDC が実装する。テストでは差し替える）。
type Provider interface {
	AuthCodeURL(state, nonce, verifier string) string
	Exchange(ctx context.Context, code, verifier string) (string, error)
	VerifyIDToken(raw, nonce string) (Principal, error)
}

type HandlersConfig struct {
	// Provider は OIDC でログインするときに渡す。nil なら OIDC のログインは 404 にする
	Provider Provider
	// Dev が true なら、開発用ログイン（POST /api/auth/dev-login）を使えるようにする
	Dev     bool
	Store   SessionStore
	Users   UserResolver
	Allowed Allowlist
	TTL     time.Duration
	Secure  bool // Cookie に Secure を付ける（HTTPS で配信するとき）
	Now     func() time.Time
}

// Handlers は /api/auth/* を実装する。リダイレクトと Cookie を扱うので、生成したコードを使わずに書く。
type Handlers struct {
	cfg HandlersConfig
}

func NewHandlers(cfg HandlersConfig) *Handlers {
	return &Handlers{cfg: cfg}
}

func (h *Handlers) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/auth/login", h.login)
	mux.HandleFunc("GET /api/auth/callback", h.callback)
	mux.HandleFunc("POST /api/auth/dev-login", h.devLogin)
}

// login はプロバイダーのログイン画面へリダイレクトする。
func (h *Handlers) login(w http.ResponseWriter, r *http.Request) {
	if h.cfg.Provider == nil {
		http.NotFound(w, r)
		return
	}
	state, nonce, verifier := randomString(16), randomString(16), randomString(32)
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookie,
		Value:    state + "." + nonce + "." + verifier,
		Path:     "/api/auth/",
		MaxAge:   int((10 * time.Minute).Seconds()),
		HttpOnly: true,
		Secure:   h.cfg.Secure,
		// プロバイダーから戻ってくるのはトップレベルの GET なので、Lax なら送られる
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, h.cfg.Provider.AuthCodeURL(state, nonce, verifier), http.StatusFound)
}

// callback はプロバイダーから戻ってきたときに、ID トークンを検証してセッションを作り、/ に戻す。
// 失敗したときは /?login=failed（使えない人なら /?login=forbidden）に戻す。
func (h *Handlers) callback(w http.ResponseWriter, r *http.Request) {
	if h.cfg.Provider == nil {
		http.NotFound(w, r)
		return
	}
	// 途中の Cookie は 1 回だけ使う
	http.SetCookie(w, &http.Cookie{Name: stateCookie, Path: "/api/auth/", MaxAge: -1, HttpOnly: true, Secure: h.cfg.Secure})

	fail := func(result, reason string, err error) {
		slog.WarnContext(r.Context(), "login failed", "reason", reason, "err", err)
		http.Redirect(w, r, "/?login="+result, http.StatusFound)
	}
	c, err := r.Cookie(stateCookie)
	if err != nil {
		fail("failed", "state cookie is missing", err)
		return
	}
	parts := strings.Split(c.Value, ".")
	q := r.URL.Query()
	if len(parts) != 3 || q.Get("state") == "" || q.Get("state") != parts[0] {
		fail("failed", "state mismatch", nil)
		return
	}
	if e := q.Get("error"); e != "" {
		fail("failed", "provider returned an error: "+e, nil)
		return
	}
	nonce, verifier := parts[1], parts[2]
	raw, err := h.cfg.Provider.Exchange(r.Context(), q.Get("code"), verifier)
	if err != nil {
		fail("failed", "code exchange", err)
		return
	}
	p, err := h.cfg.Provider.VerifyIDToken(raw, nonce)
	if err != nil {
		fail("failed", "id token", err)
		return
	}
	if !h.cfg.Allowed.Allows(p.Email) {
		fail("forbidden", "email is not allowed", nil)
		return
	}
	if err := h.startSession(w, r, p); err != nil {
		fail("failed", "session", err)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// devLogin は開発用ログイン。メールアドレスだけでその人としてセッションを作る。
func (h *Handlers) devLogin(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Dev {
		http.NotFound(w, r)
		return
	}
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&body); err != nil {
		httpx.WriteProblem(w, http.StatusBadRequest, "invalid body")
		return
	}
	email := strings.ToLower(strings.TrimSpace(body.Email))
	local, _, ok := strings.Cut(email, "@")
	if !ok || local == "" {
		httpx.WriteProblem(w, http.StatusBadRequest, "invalid email")
		return
	}
	if err := h.startSession(w, r, Principal{Issuer: "dev", Subject: email, Email: email, Name: local}); err != nil {
		slog.ErrorContext(r.Context(), "dev login failed", "err", err)
		httpx.WriteProblem(w, http.StatusInternalServerError, "failed to log in")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) startSession(w http.ResponseWriter, r *http.Request, p Principal) error {
	user, err := h.cfg.Users.ResolveUser(r.Context(), p)
	if err != nil {
		return err
	}
	id, err := h.cfg.Store.Create(r.Context(), user, h.cfg.Now(), h.cfg.TTL)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    id,
		Path:     "/",
		MaxAge:   int(h.cfg.TTL.Seconds()),
		HttpOnly: true,
		Secure:   h.cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// Allowlist はログインできる人。"user@example.com" はその人だけ、"@example.com" はそのドメインの人を許す。
type Allowlist struct {
	emails  map[string]struct{}
	domains []string
}

func NewAllowlist(entries []string) Allowlist {
	a := Allowlist{emails: map[string]struct{}{}}
	for _, e := range entries {
		e = strings.ToLower(strings.TrimSpace(e))
		switch {
		case strings.HasPrefix(e, "@") && len(e) > 1:
			a.domains = append(a.domains, e)
		case e != "":
			a.emails[e] = struct{}{}
		}
	}
	return a
}

func (a Allowlist) Allows(email string) bool {
	email = strings.ToLower(email)
	if _, ok := a.emails[email]; ok {
		return true
	}
	for _, d := range a.domains {
		if strings.HasSuffix(email, d) {
			return true
		}
	}
	return false
}
