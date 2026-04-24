-- +goose Up
-- +goose StatementBegin
ALTER TABLE movements ADD COLUMN stock_before DECIMAL(12, 4) NOT NULL CHECK (stock_before > 0);
ALTER TABLE movements ADD COLUMN stock_after DECIMAL(12, 4) NOT NULL CHECK (stock_after > 0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE movements DROP COLUMN IF EXISTS stock_before;
ALTER TABLE movements DROP COLUMN IF EXISTS stock_after;
-- +goose StatementEnd