-- +goose Up

-- +goose StatementBegin

INSERT OR IGNORE INTO tags (id, slug, name, created_at) VALUES
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'sin-correa', 'Zona sin correa', CURRENT_TIMESTAMP),
('b2c3d4e5-f6a7-8901-bcde-f234567890ab', 'agua', 'Agua', CURRENT_TIMESTAMP),
('c3d4e5f6-a7b8-9012-cdef-34567890abcd', 'sombra', 'Sombra', CURRENT_TIMESTAMP),
('d4e5f6a7-b8c9-0123-defa-4567890abcde', 'vallado', 'Vallado', CURRENT_TIMESTAMP),
('e5a6b7c8-c9d0-1234-efab-567890abcdef', 'bolsas', 'Bolsas', CURRENT_TIMESTAMP),
('f6a7b8c9-d0e1-2345-fabc-67890abcdef1', 'terraza-perros', 'Terraza que admite perros', CURRENT_TIMESTAMP);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DELETE FROM tags
WHERE slug IN ('sin-correa', 'agua', 'sombra', 'vallado', 'bolsas', 'terraza-perros');
-- +goose StatementEnd

