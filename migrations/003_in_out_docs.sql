-- +goose Up
-- +goose StatementBegin

-- Приходная накладная
CREATE TABLE incoming_docs (
    id SERIAL PRIMARY KEY,
    counterparty_id INT NOT NULL REFERENCES counterparties(id),
    warehouse_id INT NOT NULL REFERENCES warehouses(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE incoming_items (
    id SERIAL PRIMARY KEY,
    document_id INT NOT NULL REFERENCES incoming_docs(id) ON DELETE CASCADE,
    product_id INT NOT NULL REFERENCES products(id),
    quantity DECIMAL(15,6) NOT NULL CHECK (quantity > 0),
    price DECIMAL(15,2) NOT NULL CHECK (price >= 0)
);

-- Расходная накладная (отгрузка)
CREATE TABLE outgoing_docs (
    id SERIAL PRIMARY KEY,
    counterparty_id INT NOT NULL REFERENCES counterparties(id),
    warehouse_id INT NOT NULL REFERENCES warehouses(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE outgoing_items (
    id SERIAL PRIMARY KEY,
    document_id INT NOT NULL REFERENCES outgoing_docs(id) ON DELETE CASCADE,
    product_id INT NOT NULL REFERENCES products(id),
    quantity DECIMAL(15,6) NOT NULL CHECK (quantity > 0),
    price DECIMAL(15,2) NOT NULL CHECK (price >= 0),     -- продажная цена
    unit_cost DECIMAL(15,2)                              -- себестоимость (FIFO)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS outgoing_items;
DROP TABLE IF EXISTS outgoing_docs;
DROP TABLE IF EXISTS incoming_items;
DROP TABLE IF EXISTS incoming_docs;
-- +goose StatementEnd