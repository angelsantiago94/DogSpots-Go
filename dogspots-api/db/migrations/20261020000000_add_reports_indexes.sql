-- +goose Up
-- +goose StatementBegin

-- Evitar denuncias duplicadas del mismo usuario sobre el mismo contenido.
-- Se conserva la denuncia más antigua si ya existieran duplicados.
DELETE FROM reports
WHERE place_id IS NOT NULL
  AND rowid NOT IN (
    SELECT MIN(rowid) FROM reports WHERE place_id IS NOT NULL GROUP BY user_id, place_id
  );

DELETE FROM reports
WHERE comment_id IS NOT NULL
  AND rowid NOT IN (
    SELECT MIN(rowid) FROM reports WHERE comment_id IS NOT NULL GROUP BY user_id, comment_id
  );

CREATE UNIQUE INDEX IF NOT EXISTS idx_reports_user_place
    ON reports(user_id, place_id) WHERE place_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_reports_user_comment
    ON reports(user_id, comment_id) WHERE comment_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_reports_created_at ON reports(created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_reports_created_at;
DROP INDEX IF EXISTS idx_reports_user_comment;
DROP INDEX IF EXISTS idx_reports_user_place;
-- +goose StatementEnd
