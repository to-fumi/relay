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

	verifier, err := newVerifier(ctx, cfg)
	if err != nil {
		return err
	}

	// identity
	identitySvc := identityapp.NewService(identitypg.NewRepository(pool), cfg.AdminEmails)

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
	mux.Handle("/api/", chain(apiHandler,
		auth.Middleware(verifier),
		identityhttp.UserMiddleware(identitySvc),
	))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	if cfg.StaticDir != "" {
		mux.Handle("/", httpx.SPAHandler(cfg.StaticDir))
	}

	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
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

func newVerifier(ctx context.Context, cfg config.Config) (auth.Verifier, error) {
	if cfg.AuthMode == config.AuthModeEntra {
		return auth.NewEntraVerifier(ctx, cfg.EntraTenantID, cfg.EntraAudience)
	}
	slog.Warn("AUTH_MODE=dev: using development login; do not use in production")
	return auth.DevVerifier{}, nil
}

// chain は middlewares を外側から順に適用する。
func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
