package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"relay/internal/platform/httpx"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SessionCookie はセッションの ID を入れる Cookie の名前。
const SessionCookie = "relay_session"

// ErrNoSession は、セッションがない（期限切れや不明な ID）ことを表す。
var ErrNoSession = errors.New("no session")

// SessionStore はログインのセッションを保存する。Cookie にはランダムな ID だけを入れ、DB にはそのハッシュを置く。
type SessionStore interface {
	// Create は user のセッションを作り、Cookie に入れる ID を返す。
	Create(ctx context.Context, user uuid.UUID, now time.Time, ttl time.Duration) (string, error)
	// Lookup は ID のセッションのユーザーを返す。期限切れや不明な ID は ErrNoSession を返す。
	Lookup(ctx context.Context, id string, now time.Time) (uuid.UUID, error)
}

type PostgresSessionStore struct {
	pool *pgxpool.Pool
}

func NewPostgresSessionStore(pool *pgxpool.Pool) *PostgresSessionStore {
	return &PostgresSessionStore{pool: pool}
}

func (s *PostgresSessionStore) Create(ctx context.Context, user uuid.UUID, now time.Time, ttl time.Duration) (string, error) {
	id := randomString(32)
	// 期限切れのものは、作るついでに消す
	if _, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at <= $1`, now); err != nil {
		return "", err
	}
	_, err := s.pool.Exec(ctx,
		`INSERT INTO sessions (id_hash, user_id, created_at, expires_at) VALUES ($1, $2, $3, $4)`,
		hashID(id), user, now, now.Add(ttl),
	)
	return id, err
}

func (s *PostgresSessionStore) Lookup(ctx context.Context, id string, now time.Time) (uuid.UUID, error) {
	var user uuid.UUID
	err := s.pool.QueryRow(ctx,
		`SELECT user_id FROM sessions WHERE id_hash = $1 AND expires_at > $2`, hashID(id), now,
	).Scan(&user)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrNoSession
	}
	return user, err
}

// Middleware はセッションの Cookie から利用者を決め、ユーザー ID をコンテキストに入れる。なければ 401 にする。
func Middleware(store SessionStore, now func() time.Time) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(SessionCookie)
			if err != nil || c.Value == "" {
				httpx.WriteProblem(w, http.StatusUnauthorized, "not logged in")
				return
			}
			user, err := store.Lookup(r.Context(), c.Value, now())
			switch {
			case errors.Is(err, ErrNoSession):
				httpx.WriteProblem(w, http.StatusUnauthorized, "session expired")
				return
			case err != nil:
				httpx.WriteProblem(w, http.StatusInternalServerError, "failed to look up session")
				return
			}
			next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), user)))
		})
	}
}

func hashID(id string) []byte {
	sum := sha256.Sum256([]byte(id))
	return sum[:]
}

// randomString は n バイトの乱数を URL で使える Base64 にする。
func randomString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b) // crypto/rand.Read は失敗しない
	return base64.RawURLEncoding.EncodeToString(b)
}
