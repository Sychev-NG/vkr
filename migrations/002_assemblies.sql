-- +goose Up
-- +goose StatementBegin

CREATE TABLE assemblies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    output_product_id INT NOT NULL REFERENCES products(id),
    output_quantity DECIMAL(15,6) DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE assembly_components (
    id SERIAL PRIMARY KEY,
    assembly_id INT NOT NULL REFERENCES assemblies(id) ON DELETE CASCADE,
    product_id INT NOT NULL REFERENCES products(id),
    quantity DECIMAL(15,6) NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS assembly_components;
DROP TABLE IF EXISTS assemblies;
-- +goose StatementEnd