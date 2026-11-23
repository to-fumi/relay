// Package auth はログイン（OIDC と開発用ログイン）と、Cookie のセッションを扱う。
// どのモジュールにも依存しない。ユーザーの登録は UserResolver として main.go から注入する。
package auth

import (
	"context"

	"github.com/google/uuid"
)

// Principal はログインで確かめた利用者の情報。Issuer と Subject の組で人を見分ける。
type Principal struct {
	Issuer  string // ID トークンの iss（開発用ログインでは "dev"）
	Subject string // ID トークンの sub（開発用ログインではメールアドレス）
	Email   string
	Name    string
}

// UserResolver はログインした利用者をアプリのユーザーとして登録（または更新）し、ユーザー ID を返す。
type UserResolver interface {
	ResolveUser(ctx context.Context, p Principal) (uuid.UUID, error)
}

type userIDKey struct{}

// WithUserID はアプリ内のユーザー ID をコンテキストに入れる。
func WithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey{}, id)
}

func UserIDFrom(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey{}).(uuid.UUID)
	return id, ok
}
