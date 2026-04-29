-- +goose Up
-- +goose StatementBegin

-- Текущие остатки
CREATE TABLE stocks (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL REFERENCES products(id),
    warehouse_id INT NOT NULL REFERENCES warehouses(id),
    quantity DECIMAL(15,6) NOT NULL CHECK (quantity >= 0),
    UNIQUE(product_id, warehouse_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS stocks;
-- +goose StatementEnd