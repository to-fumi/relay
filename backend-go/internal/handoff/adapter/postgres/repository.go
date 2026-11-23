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

const (
	uniqueViolation = "23505"
	columns         = `id, jti, from_user, to_user, item, qty, created_at, due_date, return_requested_at, returned_at`
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, h domain.Handoff) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO handoffs (`+columns+`)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		h.ID(), h.JTI(), h.From(), h.To(), string(h.Item()), int(h.Qty()), h.CreatedAt(),
		dueDateValue(h.DueDate()), h.ReturnRequestedAt(), h.ReturnedAt(),
	)
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolation && pgErr.ConstraintName == "handoffs_jti_key" {
		return domain.ErrAlreadyRecorded
	}
	return err
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (domain.Handoff, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+columns+` FROM handoffs WHERE id = $1`, id)
	if err != nil {
		return domain.Handoff{}, err
	}
	h, err := pgx.CollectExactlyOneRow(rows, scanHandoff)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Handoff{}, domain.ErrNotFound
	}
	return h, err
}

func (r *Repository) SaveReturn(ctx context.Context, h domain.Handoff) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE handoffs SET returned_at = $2 WHERE id = $1 AND returned_at IS NULL`,
		h.ID(), h.ReturnedAt(),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAlreadyReturned
	}
	return nil
}

func (r *Repository) GetMany(ctx context.Context, ids []uuid.UUID) ([]domain.Handoff, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+columns+` FROM handoffs WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, err
	}
	hs, err := pgx.CollectRows(rows, scanHandoff)
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]domain.Handoff, len(hs))
	for _, h := range hs {
		byID[h.ID()] = h
	}
	out := make([]domain.Handoff, len(ids))
	for i, id := range ids {
		h, ok := byID[id]
		if !ok {
			return nil, domain.ErrNotFound
		}
		out[i] = h
	}
	return out, nil
}

func (r *Repository) SaveReturnRequests(ctx context.Context, hs []domain.Handoff) error {
	return r.updateAll(ctx, hs,
		`UPDATE handoffs SET return_requested_at = $2
		 WHERE id = $1 AND returned_at IS NULL AND return_requested_at IS NULL`,
		func(h domain.Handoff) *time.Time { return h.ReturnRequestedAt() },
	)
}

func (r *Repository) SaveConfirmedReturns(ctx context.Context, hs []domain.Handoff) error {
	return r.updateAll(ctx, hs,
		`UPDATE handoffs SET returned_at = $2
		 WHERE id = $1 AND returned_at IS NULL AND return_requested_at IS NOT NULL`,
		func(h domain.Handoff) *time.Time { return h.ReturnedAt() },
	)
}

// updateAll は hs のそれぞれに sql（$1 = id、$2 = value(h)）を 1 つのトランザクションで実行する。
// 1 件でも更新されなかった（条件に合わなくなっていた）場合は、すべて取り消して ErrAlreadyReturned を返す。
func (r *Repository) updateAll(ctx context.Context, hs []domain.Handoff, sql string, value func(domain.Handoff) *time.Time) error {
	return pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		for _, h := range hs {
			tag, err := tx.Exec(ctx, sql, h.ID(), value(h))
			if err != nil {
				return err
			}
			if tag.RowsAffected() == 0 {
				return domain.ErrAlreadyReturned
			}
		}
		return nil
	})
}

func (r *Repository) ListByUser(ctx context.Context, user uuid.UUID, limit int) ([]domain.Handoff, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+columns+`
		 FROM handoffs
		 WHERE from_user = $1 OR to_user = $1
		 ORDER BY created_at DESC
		 LIMIT $2`,
		user, limit,
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, scanHandoff)
}

func (r *Repository) ListOutstandingByUser(ctx context.Context, user uuid.UUID) ([]domain.Handoff, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+columns+`
		 FROM handoffs
		 WHERE (from_user = $1 OR to_user = $1) AND due_date IS NOT NULL AND returned_at IS NULL
		 ORDER BY due_date, created_at`,
		user,
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, scanHandoff)
}

func scanHandoff(row pgx.CollectableRow) (domain.Handoff, error) {
	var (
		id, from, to uuid.UUID
		jti, item    string
		qty          int
		at           time.Time
		due          *time.Time
		requestedAt  *time.Time
		returnedAt   *time.Time
	)
	if err := row.Scan(&id, &jti, &from, &to, &item, &qty, &at, &due, &requestedAt, &returnedAt); err != nil {
		return domain.Handoff{}, err
	}
	var dueDate *domain.DueDate
	if due != nil {
		d := domain.DueDateOf(*due)
		dueDate = &d
	}
	return domain.Restore(id, jti, from, to, item, qty, at, dueDate, requestedAt, returnedAt), nil
}

// dueDateValue は返却期限を date 列に入れる値にする（pgx は time.Time の日付部分を使う）。
func dueDateValue(d *domain.DueDate) *time.Time {
	if d == nil {
		return nil
	}
	t := d.Time()
	return &t
}
