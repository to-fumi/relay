// Package postgres は identity のリポジトリを PostgreSQL で実装する。
package postgres

import (
	"context"
	"errors"

	"relay/internal/identity/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const userColumns = `id, oidc_issuer, oidc_sub, display_name, email, is_admin`

func scanUser(row pgx.Row) (domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Issuer, &u.Subject, &u.DisplayName, &u.Email, &u.IsAdmin)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	return u, err
}

func (r *Repository) UpsertBySubject(ctx context.Context, u domain.User) (domain.User, error) {
	return scanUser(r.pool.QueryRow(ctx,
		`INSERT INTO users (id, oidc_issuer, oidc_sub, display_name, email, is_admin)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (oidc_issuer, oidc_sub) DO UPDATE
		 SET display_name = EXCLUDED.display_name,
		     email        = EXCLUDED.email,
		     is_admin     = users.is_admin OR EXCLUDED.is_admin
		 RETURNING `+userColumns,
		u.ID, u.Issuer, u.Subject, u.DisplayName, u.Email, u.IsAdmin,
	))
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	return scanUser(r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id))
}

func (r *Repository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+userColumns+` FROM users WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.User, error) { return scanUser(row) })
}
