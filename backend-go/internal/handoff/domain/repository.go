package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	// ErrAlreadyRecorded は同じ jti の受け渡しがすでに記録されていることを表す。
	ErrAlreadyRecorded = errors.New("handoff already recorded")
	// ErrNotFound は受け渡しの記録が見つからないことを表す。
	ErrNotFound = errors.New("handoff not found")
)

type Repository interface {
	// Save は受け渡しを記録する。jti が重複する場合は ErrAlreadyRecorded を返す。
	Save(ctx context.Context, h Handoff) error
	// Get は id の受け渡しを返す。見つからない場合は ErrNotFound を返す。
	Get(ctx context.Context, id uuid.UUID) (Handoff, error)
	// SaveReturn は h の返却日時を記録する。すでに返却済みの場合は ErrAlreadyReturned を返す。
	SaveReturn(ctx context.Context, h Handoff) error
	// GetMany は ids の受け渡しを ids の順に返す。1 件でも見つからない場合は ErrNotFound を返す。
	GetMany(ctx context.Context, ids []uuid.UUID) ([]Handoff, error)
	// SaveReturnRequests は hs の返却を送った日時をまとめて記録する。
	// 1 件でも返却済みか承認待ちになっていた場合は、どれも記録せずに ErrAlreadyReturned を返す。
	SaveReturnRequests(ctx context.Context, hs []Handoff) error
	// SaveConfirmedReturns は hs の返却日時をまとめて記録する。
	// 1 件でも返却済みか承認待ちでなくなっていた場合は、どれも記録せずに ErrAlreadyReturned を返す。
	SaveConfirmedReturns(ctx context.Context, hs []Handoff) error
	// ListByUser は user が渡した、または受け取った記録を新しい順に返す。
	ListByUser(ctx context.Context, user uuid.UUID, limit int) ([]Handoff, error)
	// ListOutstandingByUser は user が渡した、または受け取った記録のうち、
	// 返却期限があってまだ返却されていないものを、期限が近い順に返す。
	ListOutstandingByUser(ctx context.Context, user uuid.UUID) ([]Handoff, error)
}
