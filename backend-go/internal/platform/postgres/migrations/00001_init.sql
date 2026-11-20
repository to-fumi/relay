-- +goose Up
CREATE TABLE users (
    id           uuid PRIMARY KEY,
    entra_oid    text NOT NULL UNIQUE,
    display_name text NOT NULL,
    email        text NOT NULL,
    is_admin     boolean NOT NULL DEFAULT false,
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE groups (
    id         uuid PRIMARY KEY,
    name       text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE user_groups (
    user_id  uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    group_id uuid NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, group_id)
);

CREATE TABLE handoffs (
    id         uuid PRIMARY KEY,
    jti        text NOT NULL UNIQUE,
    from_user  uuid NOT NULL REFERENCES users (id),
    to_user    uuid NOT NULL REFERENCES users (id),
    item       text NOT NULL,
    qty        integer NOT NULL CHECK (qty > 0),
    created_at timestamptz NOT NULL,
    CHECK (from_user <> to_user)
);

CREATE INDEX handoffs_from_user_created_at_idx ON handoffs (from_user, created_at DESC);
CREATE INDEX handoffs_to_user_created_at_idx ON handoffs (to_user, created_at DESC);

-- +goose Down
DROP TABLE handoffs;
DROP TABLE user_groups;
DROP TABLE groups;
DROP TABLE users;
