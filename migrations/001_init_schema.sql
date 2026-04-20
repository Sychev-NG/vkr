-- +goose Up
-- +goose StatementBegin
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(8) NOT NULL CHECK (type IN ('raw', 'finished')),
    unit VARCHAR(8) NOT NULL CHECK (unit IN ('kg'))
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS products;
-- +goose StatementEnd