-- +goose Up
-- +goose StatementBegin

-- Заказ на сборку (складская операция)
CREATE TABLE assembly_orders (
    id SERIAL PRIMARY KEY,
    assembly_id INT NOT NULL REFERENCES assemblies(id),
    warehouse_id INT NOT NULL REFERENCES warehouses(id),
    quantity_to_build DECIMAL(15,6) NOT NULL CHECK (quantity_to_build > 0),
    doc_date TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Списание компонентов (автоматически из спецификации)
CREATE TABLE assembly_order_consumptions (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL REFERENCES assembly_orders(id) ON DELETE CASCADE,
    product_id INT NOT NULL REFERENCES products(id),
    quantity DECIMAL(15,6) NOT NULL CHECK (quantity > 0),
    unit_cost DECIMAL(15,2) NOT NULL,
    total_cost DECIMAL(15,2) NOT NULL
);

-- Оприходование собранного товара
CREATE TABLE assembly_order_outputs (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL REFERENCES assembly_orders(id) ON DELETE CASCADE,
    product_id INT NOT NULL REFERENCES products(id),
    quantity DECIMAL(15,6) NOT NULL CHECK (quantity > 0),
    unit_cost DECIMAL(15,2) NOT NULL,
    total_cost DECIMAL(15,2) NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS assembly_order_outputs;
DROP TABLE IF EXISTS assembly_order_consumptions;
DROP TABLE IF EXISTS assembly_orders;
-- +goose StatementEnd