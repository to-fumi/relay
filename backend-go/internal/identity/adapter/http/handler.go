// Package http は identity の API（api.StrictServerInterface の一部）と、ユーザーを解決するミドルウェアを実装する。
package http

import (
	"context"
	"errors"
	"net/http"

	"relay/internal/api"
	"relay/internal/identity/application"
	"relay/internal/platform/auth"
	"relay/internal/platform/httpx"
)

type Handler struct {
	svc *application.Service
}

func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetMe(ctx context.Context, _ api.GetMeRequestObject) (api.GetMeResponseObject, error) {
	id, ok := auth.UserIDFrom(ctx)
	if !ok {
		return nil, errors.New("user is not set in context")
	}
	u, err := h.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return api.GetMe200JSONResponse{Id: u.ID, DisplayName: u.DisplayName, Email: u.Email, IsAdmin: u.IsAdmin}, nil
}

// UserMiddleware は認証済みの Principal をユーザーとして登録し、ユーザーIDをコンテキストに入れる。
// auth.Middleware の後ろに置く。
func UserMiddleware(svc *application.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := auth.PrincipalFrom(r.Context())
			if !ok {
				httpx.WriteProblem(w, http.StatusUnauthorized, "unauthenticated")
				return
			}
			u, err := svc.EnsureUser(r.Context(), application.Identity{Subject: p.Subject, Email: p.Email, Name: p.Name})
			if err != nil {
				httpx.WriteProblem(w, http.StatusInternalServerError, "failed to resolve user")
				return
			}
			next.ServeHTTP(w, r.WithContext(auth.WithUserID(r.Context(), u.ID)))
		})
	}
}
