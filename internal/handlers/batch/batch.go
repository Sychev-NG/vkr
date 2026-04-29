package batch

import (
	"context"
	"net/http"
	"strconv"
	"time"
	"vkr/internal/entity"

	"github.com/gin-gonic/gin"
)

type BatchDTO struct {
	ID                int     `json:"id"`
	ProductID         int     `json:"product_id"`
	ProductName       string  `json:"product_name"`
	WarehouseID       int     `json:"warehouse_id"`
	WarehouseName     string  `json:"warehouse_name"`
	QuantityRemaining float64 `json:"quantity_remaining"`
	CreatedAt         string  `json:"created_at"`
}

type BatchProvider interface {
	GetByFilter(ctx context.Context, filter entity.BatchFilter) ([]entity.Batch, error)	
	GetByID(ctx context.Context, id int) (*entity.Batch, error)
}

type ProductRepository interface {
	GetByID(ctx context.Context, id int) (*entity.Product, error)
}

type WarehouseRepository interface {
	GetByID(ctx context.Context, id int) (*entity.Warehouse, error)
}

type BatchHandler struct {
	provider       BatchProvider
	productRepo   ProductRepository
	warehouseRepo WarehouseRepository
}

func New(bp BatchProvider, pr ProductRepository, wr WarehouseRepository) *BatchHandler {
	return &BatchHandler{
		provider:       bp,
		productRepo:   pr,
		warehouseRepo: wr,
	}
}

func (h *BatchHandler) List(c *gin.Context) {
	var uriParams struct {
        WarehouseID int `uri:"warehouse_id"`
        ProductID int `uri:"product_id"`
    } 
    
    if err := c.ShouldBindUri(&uriParams); err != nil {
        c.Status(http.StatusBadRequest)
        return
    }

	collection, err := h.provider.GetByFilter(c, entity.BatchFilter{WarehouseID: uriParams.WarehouseID, ProductID: uriParams.ProductID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error" : err.Error()})
		return
	}

	dtoCollection := make([]BatchDTO, len(collection))
	for index, item := range collection {

		warhouse, err := h.warehouseRepo.GetByID(c, item.WarehouseID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error" : err.Error()})
			return
		}

		product, err := h.productRepo.GetByID(c, item.ProductID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error" : err.Error()})
			return
		}

		dtoCollection[index] = BatchDTO{
			ID: item.ID,
			ProductID: item.ProductID,
			ProductName: product.Name,
			WarehouseID: item.WarehouseID,
			WarehouseName: warhouse.Name,
			QuantityRemaining: item.QuantityRemaining,
			CreatedAt: item.CreatedAt.Format(time.DateTime),
		}
	}

	c.IndentedJSON(http.StatusOK, dtoCollection)
}

func (h *BatchHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	batch, err := h.provider.GetByID(c, id)
	if err != nil {
		if err == entity.ErrBatchNotFound {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	product, err := h.productRepo.GetByID(c, batch.ProductID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	warehouse, err := h.warehouseRepo.GetByID(c, batch.WarehouseID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, BatchDTO{
		ID:                batch.ID,
		ProductID:         batch.ProductID,
		ProductName:       product.Name,
		WarehouseID:       batch.WarehouseID,
		WarehouseName:     warehouse.Name,
		QuantityRemaining: batch.QuantityRemaining,
		CreatedAt:         batch.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}