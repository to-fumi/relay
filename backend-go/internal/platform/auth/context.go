// Package auth はアクセストークンの検証と、リクエストの認証情報を扱う。
package auth

import (
	"context"

	"github.com/google/uuid"
)

// Principal は検証済みのアクセストークンから取り出した利用者の情報。
type Principal struct {
	Subject string // Entra ID の oid（dev モードでは "dev:<email>"）
	Email   string
	Name    string
}

type principalKey struct{}

type userIDKey struct{}

func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

func PrincipalFrom(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(Principal)
	return p, ok
}

// WithUserID はアプリ内のユーザーIDをコンテキストに入れる（identity モジュールが設定する）。
func WithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey{}, id)
}

func UserIDFrom(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey{}).(uuid.UUID)
	return id, ok
}
