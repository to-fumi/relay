// Package application は handoff モジュールのユースケース。domain だけに依存する。
package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid handoff token")
	ErrTokenExpired = errors.New("handoff token expired")
)

// TokenClaims は QR トークンに入れる内容。
type TokenClaims struct {
	JTI       string
	From      uuid.UUID
	Item      string
	Qty       int
	ExpiresAt time.Time
}

// TokenSigner は QR トークンの署名と検証を行う。
type TokenSigner interface {
	Sign(c TokenClaims) (string, error)
	// Verify は署名と有効期限を検証する。失敗したときは ErrInvalidToken か ErrTokenExpired を返す。
	Verify(token string, now time.Time) (TokenClaims, error)
}

// UserDirectory はユーザーの表示名を引く（identity モジュールが実装を提供する）。
type UserDirectory interface {
	DisplayNames(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error)
}
