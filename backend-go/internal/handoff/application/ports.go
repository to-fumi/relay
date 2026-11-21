// Package application は handoff モジュールのユースケース。domain だけに依存する。
package application

import (
	"context"
	"errors"
	"time"

	"relay/internal/handoff/domain"

	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid handoff token")
	ErrTokenExpired = errors.New("handoff token expired")
	// ErrReturnToken は、受け渡し用のトークンを期待したところに返却用のトークンが渡されたことを表す。
	ErrReturnToken = errors.New("this is a return token")
)

// TokenClaims は受け渡し用の QR トークンに入れる内容。
type TokenClaims struct {
	JTI       string
	From      uuid.UUID
	Item      string
	Qty       int
	DueDate   *domain.DueDate // 返却期限。なければ nil
	ExpiresAt time.Time
}

// ReturnClaims は返却用の QR トークンに入れる内容。
type ReturnClaims struct {
	HandoffID uuid.UUID
	ExpiresAt time.Time
}

// TokenSigner は QR トークンの署名と検証を行う。
type TokenSigner interface {
	Sign(c TokenClaims) (string, error)
	// Verify は受け渡し用のトークンの署名と有効期限を検証する。
	// 失敗したときは ErrInvalidToken か ErrTokenExpired を、返却用のトークンだったときは ErrReturnToken を返す。
	Verify(token string, now time.Time) (TokenClaims, error)
	SignReturn(c ReturnClaims) (string, error)
	// VerifyReturn は返却用のトークンの署名と有効期限を検証する。失敗したときは ErrInvalidToken か ErrTokenExpired を返す。
	VerifyReturn(token string, now time.Time) (ReturnClaims, error)
}

// UserDirectory はユーザーの表示名を引く（identity モジュールが実装を提供する）。
type UserDirectory interface {
	DisplayNames(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error)
}
