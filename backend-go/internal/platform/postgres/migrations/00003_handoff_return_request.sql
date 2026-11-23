-- +goose Up
-- 対面でない返却を送った日時（貸した側の承認待ち）。返却期限がある受け渡しだけに入る
ALTER TABLE handoffs
    ADD COLUMN return_requested_at timestamptz,
    ADD CONSTRAINT handoffs_return_request_requires_due CHECK (return_requested_at IS NULL OR due_date IS NOT NULL);

-- +goose Down
ALTER TABLE handoffs
    DROP CONSTRAINT handoffs_return_request_requires_due,
    DROP COLUMN return_requested_at;
