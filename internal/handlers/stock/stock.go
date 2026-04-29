package stock

import (
	"context"
	"net/http"
	"vkr/internal/entity"

	"github.com/gin-gonic/gin"
)

type StockDTO struct {
	ProductID     int     `json:"product_id"`
	ProductName   string  `json:"product_name"`
	ProductUnit   string  `json:"product_unit"`
	WarehouseID   int     `json:"warehouse_id"`
	WarehouseName string  `json:"warehouse_name"`
	Quantity      float64 `json:"quantity"`
}

type StockServiceInterface interface {
	GetByFilter(ctx context.Context, filter entity.StockFilter) ([]entity.Stock, error)
}

type ProductRepository interface {
	GetByID(ctx context.Context, id int) (*entity.Product, error)
}

type WarehouseRepository interface {
	GetByID(ctx context.Context, id int) (*entity.Warehouse, error)
}

type StockHandler struct {
	service      StockServiceInterface
	productRepo  ProductRepository
	warehouseRepo WarehouseRepository
}

func New(ssi StockServiceInterface, pr ProductRepository, wr WarehouseRepository) *StockHandler {
	return &StockHandler{
		service:      ssi,
		productRepo:  pr,
		warehouseRepo: wr,
	}
}

func (h *StockHandler) List(c *gin.Context) {
    var uriParams struct {
        WarehouseID int `uri:"warehouse_id"`
        ProductID int `uri:"product_id"`
    } 
    
    if err := c.ShouldBindUri(&uriParams); err != nil {
        c.Status(http.StatusBadRequest)
        return
    }

	collection, err := h.service.GetByFilter(c, entity.StockFilter{WarehouseID: uriParams.WarehouseID, ProductID: uriParams.ProductID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error" : err.Error()})
		return
	}

	dtoCollection := make([]StockDTO, len(collection))
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

		dtoCollection[index] = StockDTO{
			ProductID: item.ProductID,
			ProductName: product.Name,
			ProductUnit: product.Unit,
			WarehouseID: item.WarehouseID,
			WarehouseName: warhouse.Name,
			Quantity: item.Quantity,
		}
	}

	c.IndentedJSON(http.StatusOK, dtoCollection)
}