-- +goose Up
-- +goose StatementBegin
CREATE TABLE movements (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL,
    warehouse_id INTEGER NOT NULL,
    document_id INTEGER NOT NULL,
    document_type VARCHAR(10) NOT NULL,
    quantity DECIMAL(12, 4) NOT NULL,
    date TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS movements;
-- +goose StatementEnd