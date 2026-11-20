package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrAlreadyRecorded は同じ jti の受け渡しがすでに記録されていることを表す。
var ErrAlreadyRecorded = errors.New("handoff already recorded")

type Repository interface {
	// Save は受け渡しを記録する。jti が重複する場合は ErrAlreadyRecorded を返す。
	Save(ctx context.Context, h Handoff) error
	// ListByUser は user が渡した、または受け取った記録を新しい順に返す。
	ListByUser(ctx context.Context, user uuid.UUID, limit int) ([]Handoff, error)
}
