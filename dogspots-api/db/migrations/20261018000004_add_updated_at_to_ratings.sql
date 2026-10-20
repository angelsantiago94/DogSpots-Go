-- +goose Up
-- +goose StatementBegin

ALTER TABLE ratings ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE ratings DROP COLUMN updated_at;

-- +goose StatementEnd
