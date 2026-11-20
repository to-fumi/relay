// Package domain はユーザーのドメインモデル。外部のパッケージには依存しない。
package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("user not found")

type User struct {
	ID          uuid.UUID
	EntraOID    string
	DisplayName string
	Email       string
	IsAdmin     bool
}

type Repository interface {
	// UpsertByEntraOID は EntraOID でユーザーを登録または更新し、保存後のユーザーを返す。
	// 既存ユーザーの ID は変えない。IsAdmin は true への変更だけを反映する。
	UpsertByEntraOID(ctx context.Context, u User) (User, error)
	FindByID(ctx context.Context, id uuid.UUID) (User, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]User, error)
}
