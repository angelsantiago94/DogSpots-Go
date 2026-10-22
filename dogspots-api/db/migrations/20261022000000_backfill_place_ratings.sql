-- +goose Up
-- +goose StatementBegin

-- Rellenar agregados denormalizados para valoraciones previas a la fase 5.
UPDATE places
SET rating_avg = (SELECT AVG(rating) FROM ratings WHERE ratings.place_id = places.id),
    rating_count = (SELECT COUNT(*) FROM ratings WHERE ratings.place_id = places.id);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
-- Sin reversión: los agregados se recalculan en cada valoración.
SELECT 1;
-- +goose StatementEnd
