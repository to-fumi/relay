package auth

import (
	"context"
	"net/http"
	"strings"

	"relay/internal/platform/httpx"
)

// Verifier はアクセストークンを検証する。
type Verifier interface {
	Verify(ctx context.Context, token string) (Principal, error)
}

// Middleware は Bearer トークンを検証し、Principal をコンテキストに入れる。
func Middleware(v Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				httpx.WriteProblem(w, http.StatusUnauthorized, "missing bearer token")
				return
			}
			p, err := v.Verify(r.Context(), token)
			if err != nil {
				httpx.WriteProblem(w, http.StatusUnauthorized, "invalid access token")
				return
			}
			next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), p)))
		})
	}
}
