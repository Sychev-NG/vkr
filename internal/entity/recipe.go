package entity

import "errors"

var (
	ErrRecipeNotFound 		= errors.New("recipe not found")
	ErrInvalidRecipeName 	= errors.New("invalid recipe name")
	ErrRecipeDuplicateFound = errors.New("recipe duplicate found")
	
	ErrFinishedProductNotFound = errors.New("finished product not found")
	ErrRawProductNotFound = errors.New("raw product not found")
	ErrInvalidRawMaterial = errors.New("invalid raw material")
	ErrInvalidFinishedMaterial = errors.New("invalid finished material")
)

type Recipe struct {
	ID       	int
	ProductID	int
	Name		string
	Ingredients []RecipeIngredient
}

type RecipeIngredient struct {
	ID				int
	RecipeID		int
	RawMaterialID	int
	QuantityPerUnit	float32
}

type UpsertRecipeVO struct {
	ProductID	int
	Name		string
	Ingredients []UpsertIngredientVO
}

type UpsertIngredientVO struct {
	RawMaterialID	int
	QuantityPerUnit	float32
}