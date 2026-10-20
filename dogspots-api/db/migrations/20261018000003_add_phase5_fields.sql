-- +goose Up
-- +goose StatementBegin

-- Añadir conteo de respuestas a comentarios para mejorar rendimiento en GET /places/{id}/comments
ALTER TABLE comments ADD COLUMN reply_count INT DEFAULT 0;

-- Añadir campos denormalizados a lugares para evitar cálculos costosos en cada consulta
ALTER TABLE places ADD COLUMN comment_count INT DEFAULT 0;
ALTER TABLE places ADD COLUMN rating_avg REAL DEFAULT 0;
ALTER TABLE places ADD COLUMN rating_count INT DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE comments DROP COLUMN reply_count;

ALTER TABLE places DROP COLUMN comment_count;
ALTER TABLE places DROP COLUMN rating_avg;
ALTER TABLE places DROP COLUMN rating_count;

-- +goose StatementEnd
