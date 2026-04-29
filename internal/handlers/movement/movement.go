package movement

import (
	"context"
	"net/http"
	"time"
	"vkr/internal/entity"

	"github.com/gin-gonic/gin"
)

type MovementDTO struct {
	ID                	int     	`json:"id"`
	BatchID             int     	`json:"batch_id"`
	ProductID         	int     	`json:"product_id"`
	ProductName       	string  	`json:"product_name"`
	WarehouseID       	int     	`json:"warehouse_id"`
	WarehouseName     	string  	`json:"warehouse_name"`
	StockMovement		float64 	`json:"stock_movement"`
	Date         		time.Time  	`json:"created_at"`
}

type MovementProvider interface {
	GetByFilter(ctx context.Context, filter entity.MovementFilter) ([]entity.Movement, error)
}

type ProductRepository interface {
	GetByID(ctx context.Context, id int) (*entity.Product, error)
}

type WarehouseRepository interface {
	GetByID(ctx context.Context, id int) (*entity.Warehouse, error)
}

type MovementHandler struct {
	provider       MovementProvider
	productRepo   ProductRepository
	warehouseRepo WarehouseRepository
}

func New(bp MovementProvider, pr ProductRepository, wr WarehouseRepository) *MovementHandler {
	return &MovementHandler{
		provider:       bp,
		productRepo:   pr,
		warehouseRepo: wr,
	}
}

func (h *MovementHandler) List(c *gin.Context) {
	var uriParams struct {
        WarehouseID int `uri:"warehouse_id"`
        ProductID int `uri:"product_id"`
    } 
    
    if err := c.ShouldBindUri(&uriParams); err != nil {
        c.Status(http.StatusBadRequest)
        return
    }

	collection, err := h.provider.GetByFilter(c, entity.MovementFilter{
		WarehouseID: uriParams.WarehouseID, 
		ProductID: uriParams.ProductID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error" : err.Error()})
		return
	}

	dtoCollection := make([]MovementDTO, len(collection))
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

		dtoCollection[index] = MovementDTO{
			ID: item.ID,
			BatchID: item.BatchID,
			ProductID: item.ProductID,
			ProductName: product.Name,
			WarehouseID: item.WarehouseID,
			WarehouseName: warhouse.Name,
			StockMovement: item.StockMovement,
			Date: item.Date,
		}
	}

	c.IndentedJSON(http.StatusOK, dtoCollection)
}
