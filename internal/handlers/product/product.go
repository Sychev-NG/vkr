package product

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"vkr/internal/entity"

	"github.com/gin-gonic/gin"
)

type ProductURIParams struct {
	ID int `uri:"id"`
}

type ProductType string

const (
	Raw      ProductType = "raw"
	Finished ProductType = "finished"
)

type Product struct {
	ID   int		 `json:"id"`
	Name string      `json:"name" binding:"required"`
	Unit string      `json:"unit" binding:"required"`
	Type ProductType `json:"type" binding:"required"` // raw/finished
}

func (p *Product) Validate() error {
	switch p.Type {
	case Raw, Finished:
		return nil
	default:
		return fmt.Errorf("expected type '%s' or '%s', got '%s'", Raw, Finished, p.Type)
	}
}

type ProductServiceInterface interface {
	Add(ctx context.Context, name, unit, typeName string) (*entity.Product, error)
	GetById(ctx context.Context, id int) (*entity.Product, error)
	GetAll(ctx context.Context) ([]entity.Product, error)
	Update(ctx context.Context, id int, name, unit, typeName string) (*entity.Product, error)
	Delete(ctx context.Context, id int) error
}

type ProductHandler struct {
	service ProductServiceInterface
}

func New(psi ProductServiceInterface) *ProductHandler {
	return &ProductHandler{psi}
}

func (ph *ProductHandler) Get(c *gin.Context) {
    var uriParams struct {
        ID int `uri:"id"`
    } 
    
    if err := c.ShouldBindUri(&uriParams); err != nil {
        c.Status(http.StatusBadRequest)
        return
    }

	fmt.Println(uriParams.ID)

	item, err := ph.service.GetById(c, uriParams.ID)

	if err != nil {
		if errors.Is(err, entity.ErrProductNotFound) {
			c.Status(http.StatusNotFound)
			return			
		}

		c.Status(http.StatusInternalServerError)
		return
	}
	
	dto := Product{
		ID: item.ID,
		Name: item.Name,
		Unit: item.Unit,
		Type: ProductType(item.TypeName),
	}

	c.IndentedJSON(http.StatusOK, dto)
}

func (ph *ProductHandler) List(c *gin.Context) {
	items, err := ph.service.GetAll(c)
	
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	dtoCollection := make([]Product,0,len(items))

	for _, item := range items {
		dtoCollection = append(dtoCollection, Product{
			ID: item.ID,
			Name: item.Name,
			Unit: item.Unit,
			Type: ProductType(item.TypeName),
		})
	}

	c.IndentedJSON(http.StatusOK, dtoCollection)
}

func (ph *ProductHandler) Update(c *gin.Context) {
	var product Product

	if err  := c.BindJSON(&product); err != nil {
		return
	}

	if err := product.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// c.IndentedJSON(http.StatusOK, products[len(products)-1])
}

func (ph *ProductHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	fmt.Println(id)

	c.Status(http.StatusNoContent)
}

func (ph *ProductHandler) Create(c *gin.Context) {
	var product Product

	if err  := c.BindJSON(&product); err != nil {
		return
	}

	if err := product.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// c.IndentedJSON(http.StatusOK, products[len(products)-1])
}