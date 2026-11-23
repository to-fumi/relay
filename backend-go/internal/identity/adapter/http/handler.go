// Package http は identity の API（api.StrictServerInterface の一部）を実装する。
package http

import (
	"context"
	"errors"

	"relay/internal/api"
	"relay/internal/identity/application"
	"relay/internal/platform/auth"
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
