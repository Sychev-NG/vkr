package recipe

import (
	"context"
	"errors"
	"log"

	"vkr/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type QueryExecutor interface {
    QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
    Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
    Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type RecipeRepository struct {
	db QueryExecutor
}

func New(db QueryExecutor) *RecipeRepository {
	return &RecipeRepository{db: db}
}

func (pr *RecipeRepository) GetById(ctx context.Context, id int) (*entity.Recipe, error) {
	var item entity.Recipe

    err := pr.db.QueryRow(ctx, "SELECT id, product_id, name FROM recipes WHERE id = $1", id).Scan(
		&item.ID, 
		&item.ProductID, 
		&item.Name, 
	)
    
    if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrRecipeNotFound
		}

		log.Printf("RecipeRepository::GetById Error - %v", err)

        return nil, err
    }

	items, err := pr.GetIngredientsByRecipeId(ctx, item.ID)
	if err != nil {
		log.Printf("RecipeRepository::GetById GetIngredientsByRecipeId Error - %v", err)
		return nil, err
	}

	item.Ingredients = items
        
    return &item, nil
}

func (pr *RecipeRepository) GetByProductId(ctx context.Context, id int) (*entity.Recipe, error) {
	var item entity.Recipe

    err := pr.db.QueryRow(ctx, "SELECT id, product_id, name FROM recipes WHERE product_id = $1", id).Scan(
		&item.ID, 
		&item.ProductID, 
		&item.Name, 
	)
    
    if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrRecipeNotFound
		}

		log.Printf("RecipeRepository::GetByProductId Error - %v", err)

        return nil, err
    }

	items, err := pr.GetIngredientsByRecipeId(ctx, item.ID)
	if err != nil {
		log.Printf("RecipeRepository::GetByProductId GetIngredientsByRecipeId Error - %v", err)
		return nil, err
	}

	item.Ingredients = items
        
    return &item, nil
}

func (pr *RecipeRepository) GetIngredientsByRecipeId(ctx context.Context, id int) ([]entity.RecipeIngredient, error) {
	var items []entity.RecipeIngredient

	rows, err := pr.db.Query(ctx, "SELECT id, recipe_id, raw_material_id, quantity_per_unit FROM recipe_ingredients WHERE recipe_id=$1", id)
	
	if err != nil {
		log.Printf("RecipeRepository::GetIngredientsByRecipeId Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.RecipeIngredient
		rows.Scan(
			&item.ID,
			&item.RecipeID,
			&item.RawMaterialID,
			&item.QuantityPerUnit,
		)
		items = append(items, item)
	}

	return items, err
}

func (pr *RecipeRepository) GetAll(ctx context.Context) ([]entity.Recipe, error) {
	var items []entity.Recipe

	rows, err := pr.db.Query(ctx, "SELECT id, product_id, name FROM recipes")
	if err != nil {
		log.Printf("RecipeRepository::GetAll Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Recipe
		rows.Scan(
			&item.ID, 
			&item.ProductID, 
			&item.Name, 
		)

		ingredients, err := pr.GetIngredientsByRecipeId(ctx, item.ID)
		if err != nil {
			log.Printf("RecipeRepository::GetAll GetIngredientsByRecipeId Error - %v", err)
			return nil, err
		}

		item.Ingredients = ingredients

		items = append(items, item)
	}

	return items, err
}

func (pr *RecipeRepository) Add(ctx context.Context, vo entity.UpsertRecipeVO) (*entity.Recipe, error) {
	item, err := pr.GetByProductId(ctx, vo.ProductID)
	if err != nil && !errors.Is(err, entity.ErrRecipeNotFound) {
		log.Printf("RecipeRepository::Add Error - %v", err)
		return nil, err
	}

	if item != nil {
		return nil, entity.ErrRecipeDuplicateFound
	}

	recipeId, err := pr.addReceipe(ctx, vo.Name, vo.ProductID)
	if err != nil {
		log.Printf("RecipeRepository::Add Create Receipt Error - %v", err)
		return nil, err
	}

	for _, ing := range vo.Ingredients {
		_, err = pr.addReceipeIngredient(ctx, recipeId, ing.RawMaterialID, ing.QuantityPerUnit)
		if err != nil {
			log.Printf("RecipeRepository::Add Create Receipt Ingredient Error - %v", err)
			return nil, err
		}
	}

	result, err := pr.GetById(ctx, recipeId)
	if err != nil {
		log.Printf("RecipeRepository::Add GetById Error - %v", err)
		return nil, err
	}

	return result, nil
}

func (pr *RecipeRepository) addReceipe(ctx context.Context, name string, product_id int) (int, error) {
	var recipeId int
	err := pr.db.QueryRow(
		ctx, 
		"INSERT INTO recipes (name, product_id) VALUES ($1, $2) RETURNING id", 
		name, 
		product_id, 
	).Scan(
		&recipeId,
	)

	if err != nil {
		log.Printf("RecipeRepository::addReceipe Error - %v", err)
		return 0, err
	}

	return recipeId, nil
}

func (pr *RecipeRepository) addReceipeIngredient(ctx context.Context, recipe_id, raw_material_id int, quantity_per_unit float32) (int, error) {
	var ingredientId int
	err := pr.db.QueryRow(
		ctx, 
		"INSERT INTO recipe_ingredients (recipe_id, raw_material_id, quantity_per_unit) VALUES ($1, $2, $3) RETURNING id", 
		recipe_id, 
		raw_material_id, 
		quantity_per_unit, 
	).Scan(
		&ingredientId,
	)

	if err != nil {
		log.Printf("RecipeRepository::addReceipeIngredient Error - %v", err)
		return 0, err
	}
	return ingredientId, nil
}

func (pr *RecipeRepository) Update(ctx context.Context, id int, vo entity.UpsertRecipeVO) (*entity.Recipe, error) {
	_, err := pr.GetById(ctx, id)
	if err != nil {
		log.Printf("RecipeRepository::Update GetById Error - %v", err)
		return nil, err
	}

	err = pr.deleteIngredients(ctx, id)
	if	err != nil {
		log.Printf("RecipeRepository::Update deleteIngredients Error - %v", err)
		return nil, err
	}

	for _, ing := range vo.Ingredients {
		_, err = pr.addReceipeIngredient(ctx, id, ing.RawMaterialID, ing.QuantityPerUnit)
		if err != nil {
			log.Printf("RecipeRepository::Update addReceipeIngredient Error - %v", err)
			return nil, err
		}
	}

	err = pr.updateReceipe(ctx, id, vo.Name, vo.ProductID)
	if err != nil {
		log.Printf("RecipeRepository::Update updateReceipe Error - %v", err)
		return nil, err
	}

	result, err := pr.GetById(ctx, id)
	if err != nil {
		log.Printf("RecipeRepository::Update GetById Error - %v", err)
		return nil, err
	}

	return result, nil
}

func (pr *RecipeRepository) updateReceipe(ctx context.Context, recipe_id int, name string, product_id int) error {
	var recipeId int

	err := pr.db.QueryRow(
		ctx, 
		"UPDATE recipes SET name=$1, product_id=$2 WHERE id=$3 RETURNING id", 
		name, 
		product_id, 
		recipe_id, 
	).Scan(
		&recipeId,
	)

	if err != nil {
		log.Printf("RecipeRepository::updateReceipe Error - %v", err)
		return err
	}

	return nil
}

func (pr *RecipeRepository) Delete(ctx context.Context, id int) error {
	_, err := pr.db.Exec(ctx, "DELETE FROM recipes WHERE id=$1", id)
	if err != nil {
		log.Printf("RecipeRepository::Delete Error - %v", err)
	}
	return err
}

func (pr *RecipeRepository) deleteIngredients(ctx context.Context, recipe_id int) error {
	_, err := pr.db.Exec(ctx, "DELETE FROM recipe_ingredients WHERE recipe_id=$1", recipe_id)
	if err != nil {
		log.Printf("RecipeRepository::Delete Error - %v", err)
	}
	return err
}