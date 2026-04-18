package product

import (
	"context"
	"fmt"
	"net/http"
	"vkr/internal/entity"

	"github.com/gin-gonic/gin"
)

type ProductType string

const (
	Raw      ProductType = "raw"
	Finished ProductType = "finished"
)

type Product struct {
	ID   string      `json:"id"`
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
	Update(ctx context.Context, id int, name, unit, typeName string) (*entity.Product, error)
	Delete(ctx context.Context, id int) error
}

var products = []Product {
	{ID: "1", Name: "Мука", Unit: "кг", Type: "raw"},
}


type ProductHandler struct {
	service ProductServiceInterface
}

func New(psi ProductServiceInterface) *ProductHandler {
	return &ProductHandler{psi}
}

func (ph *ProductHandler) Get(c *gin.Context) {
	id := c.Param("id")

	fmt.Println(id)

	c.IndentedJSON(http.StatusOK, products[len(products)-1])
}

func (ph *ProductHandler) List(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, products)
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

	c.IndentedJSON(http.StatusOK, products[len(products)-1])
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

	c.IndentedJSON(http.StatusOK, products[len(products)-1])
}