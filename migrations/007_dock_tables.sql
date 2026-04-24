-- +goose Up
-- +goose StatementBegin
CREATE TABLE incoming_docs (
    id SERIAL PRIMARY KEY,
    counterparty_id INTEGER NOT NULL,
    warehouse_id INTEGER NOT NULL,
    date TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE TABLE incoming_doc_items (
    id SERIAL PRIMARY KEY,
    document_id INTEGER NOT NULL REFERENCES incoming_docs(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL,
    quantity DECIMAL(12, 4) NOT NULL CHECK (quantity > 0),
    price DECIMAL(12, 2) NOT NULL CHECK (price >= 0)
);

CREATE TABLE outgoing_docs (
    id SERIAL PRIMARY KEY,
    counterparty_id INTEGER NOT NULL,
    warehouse_id INTEGER NOT NULL,
    date TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE TABLE outgoing_doc_items (
    id SERIAL PRIMARY KEY,
    document_id INTEGER NOT NULL REFERENCES outgoing_docs(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL,
    quantity DECIMAL(12, 4) NOT NULL CHECK (quantity > 0),
    price DECIMAL(12, 2) NOT NULL CHECK (price >= 0)
);

CREATE TABLE production_docs (
    id SERIAL PRIMARY KEY,
    warehouse_id INTEGER NOT NULL,
    date TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE TABLE production_doc_items (
    id SERIAL PRIMARY KEY,
    document_id INTEGER NOT NULL REFERENCES production_docs(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL,
    quantity DECIMAL(12, 4) NOT NULL CHECK (quantity > 0),
    types VARCHAR(10) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS incoming_doc_items;
DROP TABLE IF EXISTS incoming_docs;

DROP TABLE IF EXISTS outgoing_doc_items;
DROP TABLE IF EXISTS outgoing_docs;

DROP TABLE IF EXISTS production_doc_items;
DROP TABLE IF EXISTS production_docs;
-- +goose StatementEnd