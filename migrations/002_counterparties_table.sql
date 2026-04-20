-- +goose Up
-- +goose StatementBegin
CREATE TABLE counterparties (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(8) NOT NULL CHECK (role IN ('supplier', 'buyer'))
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS counterparties;
-- +goose StatementEnd