package movement

import (
	"context"
	"net/http"
	"time"
	"vkr/internal/entity"

	"github.com/gin-gonic/gin"
)

type FilterParams struct {
    WarehouseID int    `form:"warehouse_id"`
    ProductID   int    `form:"product_id"`
}

func (fp *FilterParams) toFilter() (entity.MovementFilter) {
    return entity.MovementFilter{
        ProductID:    fp.ProductID,
        WarehouseID:  fp.WarehouseID,
    }
}

type Movement struct {
	ID				int 				`json:"id"`
	ProductID		int 				`json:"product_id"`
	ProductName		string 				`json:"product_name"`
	WarehouseID 	int 				`json:"warehouse_id"`
	WarehouseName	string 				`json:"warehouse_name"`
	DocumentID	 	int 				`json:"document_id"`
	DocumentType 	string 				`json:"document_type"`
	Quantity		float32				`json:"quantity"`
	Date			time.Time			`json:"date"`
}

type MovementServiceInterface interface {
	GetAll(ctx context.Context) ([]entity.Movement, error)
	GetByFilter(ctx context.Context, filter entity.MovementFilter) ([]entity.Movement, error)
}

type ProductProviderInterface interface {
	GetById(ctx context.Context, id int) (*entity.Product, error)
}

type WarehouseProviderInterface interface {
	GetById(ctx context.Context, id int) (*entity.Warehouse, error)
}

type MovementHandler struct {
	service MovementServiceInterface
	productProvider ProductProviderInterface
	warehouseProvider WarehouseProviderInterface
}

func New(rsi MovementServiceInterface, ppi ProductProviderInterface, wpi WarehouseProviderInterface) *MovementHandler {
	return &MovementHandler{rsi, ppi, wpi}
}

var collection = []Movement{
    {
        ID:            101,
        ProductID:     2,
        ProductName:   "Мука",
        WarehouseID:   1,
        WarehouseName: "Главный склад",
        DocumentID:    5,
        Quantity:      100.0,
        Date:          time.Date(2026, 4, 23, 10, 30, 0, 0, time.UTC),
    },
    {
        ID:            102,
        ProductID:     2,
        ProductName:   "Мука",
        WarehouseID:   1,
        WarehouseName: "Главный склад",
        DocumentID:    3,
        Quantity:      -9.0,
        Date:          time.Date(2026, 4, 23, 11, 15, 0, 0, time.UTC),
    },
}

func (ch *MovementHandler) List(c *gin.Context) {
	var uriSearchParams FilterParams
	if err := c.ShouldBindQuery(&uriSearchParams); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	items, err := ch.service.GetByFilter(c, uriSearchParams.toFilter())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dtoCollection := make([]Movement,0,len(items))

	for _, item := range items {

		product, err := ch.productProvider.GetById(c, item.ProductID)		
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}	

		warehouse, err := ch.warehouseProvider.GetById(c, item.WarehouseID)		
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}	

		dto := Movement{
			ID: item.ID,
			ProductID: item.ProductID,
			ProductName:product.Name,
			WarehouseID: item.WarehouseID,
			WarehouseName: warehouse.Name,
			DocumentID: item.DocumentID,
			DocumentType: string(item.DocumentType),
			Quantity: item.Quantity,
			Date: item.Date,
		}

		dtoCollection = append(dtoCollection, dto)
	}

	// TODO: Удалить как появятся живые данные
	// dtoCollection = collection

	c.IndentedJSON(http.StatusOK, dtoCollection)
}