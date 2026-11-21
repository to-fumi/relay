-- +goose Up
-- 返却期限（日付のみ）と返却日時。返却期限がない受け渡しは返却されない
ALTER TABLE handoffs
    ADD COLUMN due_date    date,
    ADD COLUMN returned_at timestamptz,
    ADD CONSTRAINT handoffs_returned_requires_due CHECK (returned_at IS NULL OR due_date IS NOT NULL);

-- 返却リスト（期限付きでまだ返却されていないもの）を引くための部分インデックス
CREATE INDEX handoffs_outstanding_from_idx ON handoffs (from_user, due_date)
    WHERE due_date IS NOT NULL AND returned_at IS NULL;
CREATE INDEX handoffs_outstanding_to_idx ON handoffs (to_user, due_date)
    WHERE due_date IS NOT NULL AND returned_at IS NULL;

-- +goose Down
DROP INDEX handoffs_outstanding_to_idx;
DROP INDEX handoffs_outstanding_from_idx;
ALTER TABLE handoffs
    DROP CONSTRAINT handoffs_returned_requires_due,
    DROP COLUMN returned_at,
    DROP COLUMN due_date;
