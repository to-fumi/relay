// Command relay は Relay の API サーバー。依存の組み立て（DI）はこのパッケージだけで行う。
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"relay/internal/api"
	"relay/internal/platform/auth"
	"relay/internal/platform/config"
	"relay/internal/platform/httpx"
	"relay/internal/platform/postgres"

	handoffhttp "relay/internal/handoff/adapter/http"
	handoffpg "relay/internal/handoff/adapter/postgres"
	handofftoken "relay/internal/handoff/adapter/token"
	handoffapp "relay/internal/handoff/application"
	identityhttp "relay/internal/identity/adapter/http"
	identitypg "relay/internal/identity/adapter/postgres"
	identityapp "relay/internal/identity/application"

	"github.com/google/uuid"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx)
	stop()
	if err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer pool.Close()
	if err := postgres.Migrate(ctx, pool); err != nil {
		return err
	}

	// identity
	identitySvc := identityapp.NewService(identitypg.NewRepository(pool), cfg.AdminEmails)

	// ログインとセッション（ユーザーの登録には identity のサービスを注入する）
	sessions := auth.NewPostgresSessionStore(pool)
	authHandlers, err := newAuthHandlers(ctx, cfg, sessions, userResolver{identitySvc})
	if err != nil {
		return err
	}

	// handoff（UserDirectory には identity のサービスを注入する）
	handoffSvc := handoffapp.NewService(
		handoffpg.NewRepository(pool),
		handofftoken.NewCompactSigner(cfg.HandoffTokenSecret),
		identitySvc,
		cfg.HandoffTokenTTL,
		time.Now,
		// 返却期限の「今日」は日本時間で決める（コンテナに tzdata がなくても使えるように固定のオフセットにする）
		time.FixedZone("Asia/Tokyo", 9*60*60),
	)

	srv := server{
		identityHandler: identityhttp.NewHandler(identitySvc),
		handoffHandler:  handoffhttp.NewHandler(handoffSvc),
	}

	apiHandler := api.HandlerWithOptions(
		api.NewStrictHandlerWithOptions(srv, nil, api.StrictHTTPServerOptions{
			RequestErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
				httpx.WriteProblem(w, http.StatusBadRequest, err.Error())
			},
			ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				slog.ErrorContext(r.Context(), "request failed", "path", r.URL.Path, "err", err)
				httpx.WriteProblem(w, http.StatusInternalServerError, "internal server error")
			},
		}),
		api.StdHTTPServerOptions{
			BaseURL: "/api",
			ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
				httpx.WriteProblem(w, http.StatusBadRequest, err.Error())
			},
		},
	)

	mux := http.NewServeMux()
	authHandlers.Register(mux) // /api/auth/*（セッションがなくても呼べる）
	mux.Handle("/api/", auth.Middleware(sessions, time.Now)(apiHandler))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	if cfg.StaticDir != "" {
		mux.Handle("/", httpx.SPAHandler(cfg.StaticDir))
	}

	// Cookie で認証するので、ほかのサイトから書き込みの API を呼ばせない（CSRF）。
	// ブラウザが送る Sec-Fetch-Site（なければ Origin）を見て、別のサイトからの POST などを 403 にする
	csrf := http.NewCrossOriginProtection()

	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           httpx.SecurityHeaders(csrf.Handler(mux)),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", cfg.Addr, "authMode", cfg.AuthMode)
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	}
}

func newAuthHandlers(ctx context.Context, cfg config.Config, sessions auth.SessionStore, users auth.UserResolver) (*auth.Handlers, error) {
	hc := auth.HandlersConfig{
		Store:   sessions,
		Users:   users,
		Allowed: auth.NewAllowlist(cfg.AllowedEmails),
		TTL:     cfg.SessionTTL,
		Now:     time.Now,
	}
	if cfg.AuthMode == config.AuthModeOIDC {
		provider, err := auth.NewOIDC(ctx, auth.OIDCConfig{
			Issuer:       cfg.OIDCIssuer,
			ClientID:     cfg.OIDCClientID,
			ClientSecret: cfg.OIDCClientSecret,
			RedirectURL:  cfg.OIDCRedirectURL,
		})
		if err != nil {
			return nil, err
		}
		hc.Provider = provider
		hc.Secure = strings.HasPrefix(cfg.OIDCRedirectURL, "https://")
		return auth.NewHandlers(hc), nil
	}
	slog.Warn("AUTH_MODE=dev: using development login; do not use in production")
	hc.Dev = true
	return auth.NewHandlers(hc), nil
}

// userResolver は identity のサービスを auth.UserResolver に合わせる。
type userResolver struct{ svc *identityapp.Service }

func (r userResolver) ResolveUser(ctx context.Context, p auth.Principal) (uuid.UUID, error) {
	u, err := r.svc.EnsureUser(ctx, identityapp.Identity{Issuer: p.Issuer, Subject: p.Subject, Email: p.Email, Name: p.Name})
	return u.ID, err
}
