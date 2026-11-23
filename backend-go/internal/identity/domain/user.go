// Package domain はユーザーのドメインモデル。外部のパッケージには依存しない。
package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("user not found")

type User struct {
	ID uuid.UUID
	// Issuer と Subject は、ログインした ID の発行元（ID トークンの iss）と、その中での人の ID（sub）。2 つの組で人を見分ける
	Issuer      string
	Subject     string
	DisplayName string
	Email       string
	IsAdmin     bool
}

type Repository interface {
	// UpsertBySubject は Issuer と Subject でユーザーを登録または更新し、保存後のユーザーを返す。
	// 既存ユーザーの ID は変えない。表示名とメールアドレスは新しい値にする。IsAdmin は true への変更だけを反映する。
	UpsertBySubject(ctx context.Context, u User) (User, error)
	FindByID(ctx context.Context, id uuid.UUID) (User, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]User, error)
}
