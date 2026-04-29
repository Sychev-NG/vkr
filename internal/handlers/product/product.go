package product

import (
	"context"
	"errors"
	"net/http"
	"vkr/internal/entity"

	"github.com/gin-gonic/gin"
)

type ProductDTO struct {
	ID       int     `json:"id"`
	Name     string  `json:"name" binding:"required"`
	Unit     string  `json:"unit" binding:"required"`
	MinStock float64 `json:"min_stock"`
}

type ProductServiceInterface interface {
	Add(ctx context.Context, name, unit string, minStock float64) (*entity.Product, error)
	GetByID(ctx context.Context, id int) (*entity.Product, error)
	GetAll(ctx context.Context) ([]entity.Product, error)
	Update(ctx context.Context, id int, name, unit string, minStock float64) (*entity.Product, error)
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

	item, err := ph.service.GetByID(c, uriParams.ID)
	if err != nil {
		if errors.Is(err, entity.ErrProductNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	dto := ProductDTO{
		ID:       item.ID,
		Name:     item.Name,
		Unit:     item.Unit,
		MinStock: item.MinStock,
	}

	c.JSON(http.StatusOK, dto)
}

func (ph *ProductHandler) List(c *gin.Context) {
	items, err := ph.service.GetAll(c)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	dtoCollection := make([]ProductDTO, 0, len(items))
	for _, item := range items {
		dtoCollection = append(dtoCollection, ProductDTO{
			ID:       item.ID,
			Name:     item.Name,
			Unit:     item.Unit,
			MinStock: item.MinStock,
		})
	}

	c.JSON(http.StatusOK, dtoCollection)
}

func (ph *ProductHandler) Update(c *gin.Context) {
    var uriParams struct {
        ID int `uri:"id"`
    } 
    
    if err := c.ShouldBindUri(&uriParams); err != nil {
        c.Status(http.StatusBadRequest)
        return
    }
	
	var dto ProductDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := ph.service.Update(c, uriParams.ID, dto.Name, dto.Unit, dto.MinStock)
	if err != nil {
		if errors.Is(err, entity.ErrProductNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if errors.Is(err, entity.ErrInvalidProductName) ||
			errors.Is(err, entity.ErrInvalidProductUnit) {
			c.Status(http.StatusBadRequest)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	response := ProductDTO{
		ID:       item.ID,
		Name:     item.Name,
		Unit:     item.Unit,
		MinStock: item.MinStock,
	}
	c.JSON(http.StatusOK, response)
}

func (ph *ProductHandler) Delete(c *gin.Context) {
    var uriParams struct {
        ID int `uri:"id"`
    } 
    
    if err := c.ShouldBindUri(&uriParams); err != nil {
        c.Status(http.StatusBadRequest)
        return
    }

	err := ph.service.Delete(c, uriParams.ID)
	if err != nil {
		if errors.Is(err, entity.ErrProductNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (ph *ProductHandler) Create(c *gin.Context) {
	var dto ProductDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := ph.service.Add(c, dto.Name, dto.Unit, dto.MinStock)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidProductName) ||
			errors.Is(err, entity.ErrInvalidProductUnit) {
			c.Status(http.StatusBadRequest)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	response := ProductDTO{
		ID:       item.ID,
		Name:     item.Name,
		Unit:     item.Unit,
		MinStock: item.MinStock,
	}
	c.JSON(http.StatusCreated, response)
}