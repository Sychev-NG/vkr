package recipes

import (
	"context"
	"errors"
	"net/http"
	"vkr/internal/entity"

	"github.com/gin-gonic/gin"
)

type UpsertRecipe struct {
	ProductID	int 						`json:"product_id"`
	Name 		string						`json:"name" binding:"required"`
	Ingredients []UpsertRecipeIngredient 	`json:"ingredients" binding:"required"`
}

type UpsertRecipeIngredient struct {
	RawMaterialID 	int 		`json:"raw_material_id" binding:"required"`
	QuantityPerUnit float32 	`json:"quantity_per_unit" binding:"required"`
}

func (r *UpsertRecipe) toVO() entity.UpsertRecipeVO {
	var ingredients []entity.UpsertIngredientVO

	for _, iningredient := range r.Ingredients {
		ingredients = append(ingredients, entity.UpsertIngredientVO{
			RawMaterialID: iningredient.RawMaterialID,
			QuantityPerUnit: iningredient.QuantityPerUnit,
		})
	}

	return entity.UpsertRecipeVO{
		ProductID: r.ProductID,
		Name: r.Name,
		Ingredients: ingredients,
	}
}

type Recipe struct {
	ID 			int 				`json:"id"`
	ProductID 	int 				`json:"product_id"`
	Name 		string 				`json:"name"`
	Ingredietns []RecipeIngredient 	`json:"ingredients"`
}

type RecipeIngredient struct {
	ID 			int 	`json:"id"`
	Name 		string 	`json:"name"`
	Quantity 	float32	`json:"quantity"`
	Unit 		string 	`json:"unit"`
}

type RecipeServiceInterface interface {
	Add(ctx context.Context, vo entity.UpsertRecipeVO) (*entity.Recipe, error)
	GetById(ctx context.Context, id int) (*entity.Recipe, error)
	GetAll(ctx context.Context) ([]entity.Recipe, error)
	Update(ctx context.Context, id int, vo entity.UpsertRecipeVO) (*entity.Recipe, error)
	Delete(ctx context.Context, id int) error
}

type ProductProviderInterface interface {
	GetById(ctx context.Context, id int) (*entity.Product, error)
}

type RecipeHandler struct {
	service RecipeServiceInterface
	productProvider ProductProviderInterface
}

func New(rsi RecipeServiceInterface, ppi ProductProviderInterface) *RecipeHandler {
	return &RecipeHandler{rsi, ppi}
}

func (ch *RecipeHandler) Get(c *gin.Context) {
    var uriParams struct {
        ID int `uri:"id"`
    } 
    
    if err := c.ShouldBindUri(&uriParams); err != nil {
        c.Status(http.StatusBadRequest)
        return
    }

	item, err := ch.service.GetById(c, uriParams.ID)

	if err != nil {
		if errors.Is(err, entity.ErrRecipeNotFound) {
			c.Status(http.StatusNotFound)
			return			
		}

		c.Status(http.StatusInternalServerError)
		return
	}
	
	var ingredients []RecipeIngredient

	for _, ingredient := range item.Ingredients {
		rawMaterial, err := ch.productProvider.GetById(c, ingredient.RawMaterialID)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		ingredients = append(ingredients, RecipeIngredient{
			ID: ingredient.ID,
			Name: rawMaterial.Name,
			Unit: rawMaterial.Unit,
			Quantity: ingredient.QuantityPerUnit,
		})
	}

	dto := Recipe{
		ID: item.ID,
		Name: item.Name,
		ProductID: item.ProductID,
		Ingredietns: ingredients,
	}

	c.IndentedJSON(http.StatusOK, dto)
}

func (ch *RecipeHandler) List(c *gin.Context) {
	items, err := ch.service.GetAll(c)
	
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	dtoCollection := make([]Recipe,0,len(items))

	for _, item := range items {
		var ingredients []RecipeIngredient

		for _, ingredient := range item.Ingredients {
			rawMaterial, err := ch.productProvider.GetById(c, ingredient.RawMaterialID)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}

			ingredients = append(ingredients, RecipeIngredient{
				ID: ingredient.ID,
				Name: rawMaterial.Name,
				Unit: rawMaterial.Unit,
				Quantity: ingredient.QuantityPerUnit,
			})
		}

		dto := Recipe{
			ID: item.ID,
			Name: item.Name,
			ProductID: item.ProductID,
			Ingredietns: ingredients,
		}

		dtoCollection = append(dtoCollection, dto)
	}

	c.IndentedJSON(http.StatusOK, dtoCollection)
}

func (ch *RecipeHandler) Create(c *gin.Context) {
	var recipe UpsertRecipe

	if err  := c.BindJSON(&recipe); err != nil {
		return
	}

	item, err := ch.service.Add(
		c, 
		recipe.toVO(),
	)

	if err != nil {
			httpCode := http.StatusInternalServerError

			if errors.Is(err, entity.ErrFinishedProductNotFound) || 
			   errors.Is(err, entity.ErrRawProductNotFound) ||
			   errors.Is(err, entity.ErrInvalidRawMaterial) ||
			   errors.Is(err, entity.ErrInvalidFinishedMaterial) ||
			   errors.Is(err, entity.ErrInvalidRecipeName) {
				httpCode = http.StatusBadRequest
			}

		c.JSON(httpCode, gin.H{"error": err.Error()})
		return
	}

	var ingredientsDto []RecipeIngredient

	for _, ingredient := range item.Ingredients {
		rawMaterial, err := ch.productProvider.GetById(c, ingredient.RawMaterialID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ingredientsDto = append(ingredientsDto, RecipeIngredient{
			ID: ingredient.ID,
			Name: rawMaterial.Name,
			Quantity: ingredient.QuantityPerUnit,
			Unit: rawMaterial.Unit,
		})
	}

	dto := Recipe{
		ID: item.ID,
		ProductID: item.ProductID,
		Name: item.Name,
		Ingredietns: ingredientsDto,
	}

	c.IndentedJSON(http.StatusOK, dto)
}

func (ch *RecipeHandler) Update(c *gin.Context) {
    var uriParams struct {
        ID int `uri:"id"`
    } 
    
    if err := c.ShouldBindUri(&uriParams); err != nil {
        c.Status(http.StatusBadRequest)
        return
    }

	var recipe UpsertRecipe

	if err  := c.BindJSON(&recipe); err != nil {
		return
	}

	item, err := ch.service.Update(
		c, 
		uriParams.ID,
		recipe.toVO(),
	)

	if err != nil {
		if errors.Is(err, entity.ErrRecipeDuplicateFound) {
			c.Status(http.StatusConflict)
			return
		}

		if errors.Is(err, entity.ErrRecipeNotFound) {
			c.Status(http.StatusNotFound)
			return
		}

		if errors.Is(err, entity.ErrInvalidRecipeName) {
			c.Status(http.StatusBadRequest)
			return	
		}

		c.Status(http.StatusInternalServerError)
		return
	}

	var ingredientsDto []RecipeIngredient

	for _, ingredient := range item.Ingredients {

		rawMaterial, err := ch.productProvider.GetById(c, ingredient.RawMaterialID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ingredientsDto = append(ingredientsDto, RecipeIngredient{
			ID: ingredient.ID,
			Name: rawMaterial.Name,
			Quantity: ingredient.QuantityPerUnit,
			Unit: rawMaterial.Unit,
		})
	}

	dto := Recipe{
		ID: item.ID,
		ProductID: item.ProductID,
		Name: item.Name,
		Ingredietns: ingredientsDto,
	}

	c.IndentedJSON(http.StatusOK, dto)
}

func (ch *RecipeHandler) Delete(c *gin.Context) {
    var uriParams struct {
        ID int `uri:"id"`
    } 
    
    if err := c.ShouldBindUri(&uriParams); err != nil {
        c.Status(http.StatusBadRequest)
        return
    }

	err := ch.service.Delete(c, uriParams.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}