// Package http は handoff の API（api.StrictServerInterface の一部）を実装する。
package http

import (
	"context"
	"errors"
	"net/http"

	"relay/internal/api"
	"relay/internal/handoff/application"
	"relay/internal/handoff/domain"
	"relay/internal/platform/auth"
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

var errNoUser = errors.New("user is not set in context")

type Handler struct {
	svc *application.Service
}

func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) IssueHandoffToken(ctx context.Context, req api.IssueHandoffTokenRequestObject) (api.IssueHandoffTokenResponseObject, error) {
	me, ok := auth.UserIDFrom(ctx)
	if !ok {
		return nil, errNoUser
	}
	issued, err := h.svc.IssueToken(ctx, me, req.Body.Item, req.Body.Qty)
	switch {
	case errors.Is(err, domain.ErrInvalidItem), errors.Is(err, domain.ErrInvalidQuantity):
		return api.IssueHandoffToken400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: api.BadRequestApplicationProblemPlusJSONResponse(problem(http.StatusBadRequest, err))}, nil
	case err != nil:
		return nil, err
	}
	return api.IssueHandoffToken201JSONResponse{Token: issued.Token, ExpiresAt: issued.ExpiresAt}, nil
}

func (h *Handler) ReceiveHandoff(ctx context.Context, req api.ReceiveHandoffRequestObject) (api.ReceiveHandoffResponseObject, error) {
	me, ok := auth.UserIDFrom(ctx)
	if !ok {
		return nil, errNoUser
	}
	v, err := h.svc.Receive(ctx, me, req.Body.Token)
	switch {
	case errors.Is(err, application.ErrInvalidToken), errors.Is(err, application.ErrTokenExpired):
		return api.ReceiveHandoff400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: api.BadRequestApplicationProblemPlusJSONResponse(problem(http.StatusBadRequest, err))}, nil
	case errors.Is(err, domain.ErrAlreadyRecorded):
		return api.ReceiveHandoff409ApplicationProblemPlusJSONResponse(problem(http.StatusConflict, err)), nil
	case errors.Is(err, domain.ErrSelfHandoff):
		return api.ReceiveHandoff422ApplicationProblemPlusJSONResponse(problem(http.StatusUnprocessableEntity, err)), nil
	case err != nil:
		return nil, err
	}
	return api.ReceiveHandoff201JSONResponse(toAPI(v)), nil
}

func (h *Handler) ListMyHandoffs(ctx context.Context, req api.ListMyHandoffsRequestObject) (api.ListMyHandoffsResponseObject, error) {
	me, ok := auth.UserIDFrom(ctx)
	if !ok {
		return nil, errNoUser
	}
	limit := defaultLimit
	if req.Params.Limit != nil {
		limit = min(max(*req.Params.Limit, 1), maxLimit)
	}
	vs, err := h.svc.ListMine(ctx, me, limit)
	if err != nil {
		return nil, err
	}
	out := make(api.ListMyHandoffs200JSONResponse, len(vs))
	for i, v := range vs {
		out[i] = toAPI(v)
	}
	return out, nil
}

func toAPI(v application.HandoffView) api.Handoff {
	return api.Handoff{
		Id:        v.ID,
		From:      api.UserSummary{Id: v.From.ID, DisplayName: v.From.DisplayName},
		To:        api.UserSummary{Id: v.To.ID, DisplayName: v.To.DisplayName},
		Item:      v.Item,
		Qty:       v.Qty,
		CreatedAt: v.CreatedAt,
	}
}

func problem(status int, err error) api.Problem {
	return api.Problem{Status: status, Title: err.Error()}
}
