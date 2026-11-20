// Package postgres は handoff のリポジトリを PostgreSQL で実装する。
package postgres

import (
	"context"
	"errors"
	"time"

	"relay/internal/handoff/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uniqueViolation = "23505"

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, h domain.Handoff) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO handoffs (id, jti, from_user, to_user, item, qty, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		h.ID(), h.JTI(), h.From(), h.To(), string(h.Item()), int(h.Qty()), h.CreatedAt(),
	)
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolation && pgErr.ConstraintName == "handoffs_jti_key" {
		return domain.ErrAlreadyRecorded
	}
	return err
}

func (r *Repository) ListByUser(ctx context.Context, user uuid.UUID, limit int) ([]domain.Handoff, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, jti, from_user, to_user, item, qty, created_at
		 FROM handoffs
		 WHERE from_user = $1 OR to_user = $1
		 ORDER BY created_at DESC
		 LIMIT $2`,
		user, limit,
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Handoff, error) {
		var (
			id, from, to uuid.UUID
			jti, item    string
			qty          int
			at           time.Time
		)
		if err := row.Scan(&id, &jti, &from, &to, &item, &qty, &at); err != nil {
			return domain.Handoff{}, err
		}
		return domain.Restore(id, jti, from, to, item, qty, at), nil
	})
}
