package stock

import (
	"context"
	"net/http"
	"vkr/internal/entity"

	"github.com/gin-gonic/gin"
)

type FilterParams struct {
	WarehouseID int `form:"warehouse_id"`
	ProductID   int `form:"product_id"`
}

func (fp *FilterParams) toFilter() entity.StockFilter {
	return entity.StockFilter{ProductID: fp.ProductID, WarehouseID: fp.WarehouseID}
}

type Stock struct {
	ProductID		int 				`json:"product_id"`
	ProductName		string 				`json:"product_name"`
	WarehouseID 	int 				`json:"warehouse_id"`
	WarehouseName	string 				`json:"warehouse_name"`
	Unit			string 				`json:"unit"`
	Quantity		float32				`json:"quantity"`
}

type StockServiceInterface interface {
	GetAll(ctx context.Context) ([]entity.Stock, error)
	GetByFilter(ctx context.Context, filter entity.StockFilter) ([]entity.Stock, error)
}

type ProductProviderInterface interface {
	GetById(ctx context.Context, id int) (*entity.Product, error)
}

type WarehouseProviderInterface interface {
	GetById(ctx context.Context, id int) (*entity.Warehouse, error)
}

type StockHandler struct {
	service StockServiceInterface
	productProvider ProductProviderInterface
	warehouseProvider WarehouseProviderInterface
}

func New(rsi StockServiceInterface, ppi ProductProviderInterface, wpi WarehouseProviderInterface) *StockHandler {
	return &StockHandler{rsi, ppi, wpi}
}

func (h *StockHandler) List(c *gin.Context) {
	var uriSearchParams FilterParams
	if err := c.ShouldBindQuery(&uriSearchParams); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	items, err := h.service.GetByFilter(c, uriSearchParams.toFilter())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var dtoCollection []Stock

	for _, item := range items {

		product, err := h.productProvider.GetById(c, item.ProductID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}	

		warehouse, err := h.warehouseProvider.GetById(c, item.WarehouseID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}	

		dtoCollection = append(dtoCollection, Stock{
			ProductID: item.ProductID,
			ProductName: product.Name,
			WarehouseID: item.WarehouseID,
			WarehouseName: warehouse.Name,
			Unit: product.Unit,
			Quantity: item.Quantity,
		})
	}

	if len(dtoCollection) == 0 {
		c.Status(http.StatusNoContent)
	} else {
		c.IndentedJSON(http.StatusOK, dtoCollection)
	}
}