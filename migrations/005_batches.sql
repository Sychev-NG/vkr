-- +goose Up
-- +goose StatementBegin

-- Партии товаров
CREATE TABLE batches (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL REFERENCES products(id),
    warehouse_id INT NOT NULL REFERENCES warehouses(id),
    document_id INT NOT NULL,
    document_type VARCHAR(30) NOT NULL,
    quantity_remaining DECIMAL(15,6) NOT NULL CHECK (quantity_remaining >= 0),
    unit_cost DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Движения по партиям
CREATE TABLE batch_movements (
    id SERIAL PRIMARY KEY,
    batch_id INT NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
    document_id INT NOT NULL,
    document_type VARCHAR(30) NOT NULL,
    quantity DECIMAL(15,6) NOT NULL,
    moved_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS batch_movements;
DROP TABLE IF EXISTS batches;
-- +goose StatementEnd