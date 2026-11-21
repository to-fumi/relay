// Package http は handoff の API（api.StrictServerInterface の一部）を実装する。
package http

import (
	"context"
	"errors"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"relay/internal/api"
	"relay/internal/handoff/application"
	"relay/internal/handoff/domain"
	"relay/internal/platform/auth"
)

const (
	defaultLimit = 50
	maxLimit     = 200

	// problemReturnToken は、受け取る画面で返却用の QR を読み取ったときの Problem の type。
	problemReturnToken = "urn:relay:problem:return-token"
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
	var due *domain.DueDate
	if req.Body.DueDate != nil {
		d := domain.DueDateOf(req.Body.DueDate.Time)
		due = &d
	}
	issued, err := h.svc.IssueToken(ctx, me, req.Body.Item, req.Body.Qty, due)
	switch {
	case errors.Is(err, domain.ErrInvalidItem), errors.Is(err, domain.ErrInvalidQuantity), errors.Is(err, domain.ErrDueDateInPast):
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
	case errors.Is(err, application.ErrReturnToken):
		p := problem(http.StatusBadRequest, err)
		p.Type = new(problemReturnToken)
		return api.ReceiveHandoff400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: api.BadRequestApplicationProblemPlusJSONResponse(p)}, nil
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

func (h *Handler) ListMyOutstandingHandoffs(ctx context.Context, _ api.ListMyOutstandingHandoffsRequestObject) (api.ListMyOutstandingHandoffsResponseObject, error) {
	me, ok := auth.UserIDFrom(ctx)
	if !ok {
		return nil, errNoUser
	}
	vs, err := h.svc.ListOutstanding(ctx, me)
	if err != nil {
		return nil, err
	}
	out := make(api.ListMyOutstandingHandoffs200JSONResponse, len(vs))
	for i, v := range vs {
		out[i] = toAPI(v)
	}
	return out, nil
}

func (h *Handler) IssueReturnToken(ctx context.Context, req api.IssueReturnTokenRequestObject) (api.IssueReturnTokenResponseObject, error) {
	me, ok := auth.UserIDFrom(ctx)
	if !ok {
		return nil, errNoUser
	}
	issued, err := h.svc.IssueReturnToken(ctx, me, req.Body.HandoffId)
	switch {
	// 借りた本人でない場合も、記録があることを知らせないために 404 にする
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, domain.ErrNotBorrower):
		return api.IssueReturnToken404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: api.NotFoundApplicationProblemPlusJSONResponse(problem(http.StatusNotFound, domain.ErrNotFound))}, nil
	case errors.Is(err, domain.ErrNotReturnable):
		return api.IssueReturnToken400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: api.BadRequestApplicationProblemPlusJSONResponse(problem(http.StatusBadRequest, err))}, nil
	case errors.Is(err, domain.ErrAlreadyReturned):
		return api.IssueReturnToken409ApplicationProblemPlusJSONResponse(problem(http.StatusConflict, err)), nil
	case err != nil:
		return nil, err
	}
	return api.IssueReturnToken201JSONResponse{Token: issued.Token, ExpiresAt: issued.ExpiresAt}, nil
}

func (h *Handler) ReturnHandoff(ctx context.Context, req api.ReturnHandoffRequestObject) (api.ReturnHandoffResponseObject, error) {
	me, ok := auth.UserIDFrom(ctx)
	if !ok {
		return nil, errNoUser
	}
	v, err := h.svc.Return(ctx, me, req.Body.Token)
	switch {
	case errors.Is(err, application.ErrInvalidToken), errors.Is(err, application.ErrTokenExpired), errors.Is(err, domain.ErrNotReturnable):
		return api.ReturnHandoff400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: api.BadRequestApplicationProblemPlusJSONResponse(problem(http.StatusBadRequest, err))}, nil
	case errors.Is(err, domain.ErrNotLender):
		return api.ReturnHandoff403ApplicationProblemPlusJSONResponse(problem(http.StatusForbidden, err)), nil
	case errors.Is(err, domain.ErrNotFound):
		return api.ReturnHandoff404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: api.NotFoundApplicationProblemPlusJSONResponse(problem(http.StatusNotFound, err))}, nil
	case errors.Is(err, domain.ErrAlreadyReturned):
		return api.ReturnHandoff409ApplicationProblemPlusJSONResponse(problem(http.StatusConflict, err)), nil
	case err != nil:
		return nil, err
	}
	return api.ReturnHandoff200JSONResponse(toAPI(v)), nil
}

func toAPI(v application.HandoffView) api.Handoff {
	out := api.Handoff{
		Id:         v.ID,
		From:       api.UserSummary{Id: v.From.ID, DisplayName: v.From.DisplayName},
		To:         api.UserSummary{Id: v.To.ID, DisplayName: v.To.DisplayName},
		Item:       v.Item,
		Qty:        v.Qty,
		CreatedAt:  v.CreatedAt,
		ReturnedAt: v.ReturnedAt,
	}
	if v.DueDate != nil {
		out.DueDate = &openapi_types.Date{Time: *v.DueDate}
	}
	return out
}

func problem(status int, err error) api.Problem {
	return api.Problem{Status: status, Title: err.Error()}
}
