-- +goose Up
-- +goose StatementBegin
CREATE TABLE recipes (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL
);

-- Таблица ингредиентов рецептов
CREATE TABLE recipe_ingredients (
    id SERIAL PRIMARY KEY,
    recipe_id INTEGER NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    raw_material_id INTEGER NOT NULL,
    quantity_per_unit DECIMAL(12, 4) NOT NULL CHECK (quantity_per_unit > 0)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recipes;
DROP TABLE IF EXISTS recipe_ingredients;
-- +goose StatementEnd