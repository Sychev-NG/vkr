-- +goose Up
-- +goose StatementBegin

-- Товары ()
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    unit VARCHAR(20) NOT NULL,           -- kg, piece, pack, liter
    min_stock DECIMAL(15,3) DEFAULT 0,   -- порог для алертов
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Склады
CREATE TABLE warehouses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Контрагенты (поставщики/покупатели)
CREATE TABLE counterparties (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('supplier', 'buyer')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS counterparties;
DROP TABLE IF EXISTS warehouses;
DROP TABLE IF EXISTS products;
-- +goose StatementEnd