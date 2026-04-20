package warehouses

import (
	"context"
	"errors"
	"net/http"
	"vkr/internal/entity"

	"github.com/gin-gonic/gin"
)

type Warehouse struct {
	ID   int		 `json:"id"`
	Name string      `json:"name" binding:"required"`
	Address string   `json:"address" binding:"required"`
}

type WarehouseServiceInterface interface {
	Add(ctx context.Context, name, address string) (*entity.Warehouse, error)
	GetById(ctx context.Context, id int) (*entity.Warehouse, error)
	GetAll(ctx context.Context) ([]entity.Warehouse, error)
	Update(ctx context.Context, id int, name, address string) (*entity.Warehouse, error)
	Delete(ctx context.Context, id int) error
}

type WarehouseHandler struct {
	service WarehouseServiceInterface
}

func New(csi WarehouseServiceInterface) *WarehouseHandler {
	return &WarehouseHandler{csi}
}

func (ch *WarehouseHandler) Get(c *gin.Context) {
    var uriParams struct {
        ID int `uri:"id"`
    } 
    
    if err := c.ShouldBindUri(&uriParams); err != nil {
        c.Status(http.StatusBadRequest)
        return
    }

	item, err := ch.service.GetById(c, uriParams.ID)

	if err != nil {
		if errors.Is(err, entity.ErrWarehouseNotFound) {
			c.Status(http.StatusNotFound)
			return			
		}

		c.Status(http.StatusInternalServerError)
		return
	}
	
	dto := Warehouse{
		ID: item.ID,
		Name: item.Name,
		Address: item.Address,
	}

	c.IndentedJSON(http.StatusOK, dto)
}

func (ch *WarehouseHandler) List(c *gin.Context) {
	items, err := ch.service.GetAll(c)
	
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	dtoCollection := make([]Warehouse,0,len(items))

	for _, item := range items {
		dtoCollection = append(dtoCollection, Warehouse{
			ID: item.ID,
			Name: item.Name,
			Address: item.Address,
		})
	}

	c.IndentedJSON(http.StatusOK, dtoCollection)
}

func (ch *WarehouseHandler) Update(c *gin.Context) {
    var uriParams struct {
        ID int `uri:"id"`
    } 
    
    if err := c.ShouldBindUri(&uriParams); err != nil {
        c.Status(http.StatusBadRequest)
        return
    }

	var warehouse Warehouse

	if err  := c.BindJSON(&warehouse); err != nil {
		return
	}

	item, err := ch.service.Update(
		c, 
		uriParams.ID,
		warehouse.Name,
		warehouse.Address,
	)

	if err != nil {
		if errors.Is(err, entity.ErrWarehouseDuplicateFound) {
			c.Status(http.StatusConflict)
			return
		}

		if errors.Is(err, entity.ErrWarehouseNotFound) {
			c.Status(http.StatusNotFound)
			return
		}

		if errors.Is(err, entity.ErrInvalidWarehouseAddress) ||
		   errors.Is(err, entity.ErrInvalidWarehouseName) {
			c.Status(http.StatusBadRequest)
			return	
		}

		c.Status(http.StatusInternalServerError)
		return
	}

	dto := Warehouse{
		ID: item.ID,
		Name: item.Name,
		Address: item.Address,
	}

	c.IndentedJSON(http.StatusOK, dto)
}

func (ch *WarehouseHandler) Delete(c *gin.Context) {
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

func (ch *WarehouseHandler) Create(c *gin.Context) {
	var warehouse Warehouse

	if err  := c.BindJSON(&warehouse); err != nil {
		return
	}

	item, err := ch.service.Add(
		c, 
		warehouse.Name,
		warehouse.Address,
	)

	if err != nil {
		if errors.Is(err, entity.ErrWarehouseDuplicateFound) {
			c.Status(http.StatusConflict)
			return
		}

		if errors.Is(err, entity.ErrInvalidWarehouseAddress) ||
		   errors.Is(err, entity.ErrInvalidWarehouseName) {
			c.Status(http.StatusBadRequest)
			return	
		}

		c.Status(http.StatusInternalServerError)
		return
	}

	dto := Warehouse{
		ID: item.ID,
		Name: item.Name,
		Address: item.Address,
	}

	c.IndentedJSON(http.StatusOK, dto)
}