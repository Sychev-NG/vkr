-- +goose Up
-- +goose StatementBegin
CREATE VIEW movements AS
SELECT 
    bm.id AS movement_id,
    bm.batch_id,
    bm.document_id AS movement_document_id,
    bm.document_type AS movement_document_type,
    bm.quantity,
    bm.moved_at,
    -- Информация о партии
    b.product_id,
    b.warehouse_id,
    b.unit_cost
FROM batch_movements bm INNER JOIN batches b ON bm.batch_id = b.id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS movements;
-- +goose StatementEnd