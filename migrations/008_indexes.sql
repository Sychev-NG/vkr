-- +goose Up
-- +goose StatementBegin

CREATE INDEX idx_batches_fifo ON batches(product_id, warehouse_id, created_at) WHERE quantity_remaining > 0;
CREATE INDEX idx_batches_product ON batches(product_id);
CREATE INDEX idx_batches_warehouse ON batches(warehouse_id);
CREATE INDEX idx_batch_movements_batch ON batch_movements(batch_id);
CREATE INDEX idx_stocks_lookup ON stocks(product_id, warehouse_id);
CREATE INDEX idx_alerts_unresolved ON alerts(is_resolved) WHERE is_resolved = FALSE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_alerts_unresolved;
DROP INDEX IF EXISTS idx_stocks_lookup;
DROP INDEX IF EXISTS idx_batch_movements_batch;
DROP INDEX IF EXISTS idx_batches_warehouse;
DROP INDEX IF EXISTS idx_batches_product;
DROP INDEX IF EXISTS idx_batches_fifo;
-- +goose StatementEnd