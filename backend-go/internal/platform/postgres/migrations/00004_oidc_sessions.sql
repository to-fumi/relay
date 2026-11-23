-- +goose Up
-- Entra ID の oid の代わりに、OIDC の iss と sub の組で人を見分ける
ALTER TABLE users RENAME COLUMN entra_oid TO oidc_sub;
ALTER TABLE users ADD COLUMN oidc_issuer text NOT NULL DEFAULT '';
-- 開発用ログインのユーザー（"dev:<email>"）は、iss を "dev"、sub をメールアドレスにする
UPDATE users SET oidc_issuer = 'dev', oidc_sub = substr(oidc_sub, 5) WHERE oidc_sub LIKE 'dev:%';
UPDATE users SET oidc_issuer = 'entra' WHERE oidc_issuer = '';
ALTER TABLE users ALTER COLUMN oidc_issuer DROP DEFAULT;
ALTER TABLE users DROP CONSTRAINT users_entra_oid_key;
ALTER TABLE users ADD CONSTRAINT users_oidc_issuer_sub_key UNIQUE (oidc_issuer, oidc_sub);

-- ログインのセッション。Cookie にはランダムな ID を入れ、ここにはそのハッシュ（SHA-256）を置く
CREATE TABLE sessions (
    id_hash    bytea PRIMARY KEY,
    user_id    uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL
);
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);

-- +goose Down
DROP TABLE sessions;
ALTER TABLE users DROP CONSTRAINT users_oidc_issuer_sub_key;
UPDATE users SET oidc_sub = 'dev:' || oidc_sub WHERE oidc_issuer = 'dev';
ALTER TABLE users DROP COLUMN oidc_issuer;
ALTER TABLE users RENAME COLUMN oidc_sub TO entra_oid;
ALTER TABLE users ADD CONSTRAINT users_entra_oid_key UNIQUE (entra_oid);
