package production

import (
	"context"
	"errors"
	"net/http"
	"vkr/internal/entity"
	"vkr/internal/entity/document/production"

	"github.com/gin-gonic/gin"
)

type UpsertProduction struct {
	WarehouseID		int 					`json:"warehouse_id" binding:"required"`
	Items 			[]UpsertProductionItem 	`json:"items" binding:"required"`
}

type UpsertProductionItem struct {
	FinishedMaterialID 	int 		`json:"product_id" binding:"required"`
	RecipeID 			int 		`json:"recipe_id" binding:"required"`
	Quantity 			float32		`json:"quantity" binding:"required"`
}

func (r *UpsertProduction) toVO() production.UpsertProductionDocumentVO {
	var items []production.UpsertProductionDocumentItemVO

	for _, item := range r.Items {
		items = append(items, production.UpsertProductionDocumentItemVO{
			FinishedMaterialID: item.FinishedMaterialID,
			Quantity: item.Quantity,
			RecipeID: item.RecipeID,
		})
	}

	return production.UpsertProductionDocumentVO{
		WarehouseID: r.WarehouseID,
		Items: items,
	}
}

type Production struct {
	DocumentID	int 			`json:"document_id"`
	Status 		string 			`json:"status"`
	Items 		[]ProductionItem 	`json:"movements"`
}

type ProductionItem struct {
	ProductID 	int 	`json:"product_id"`
	Quantity 	float32	`json:"quantity"`
	StockBefore float32	`json:"stock_before"`
	StockAfter 	float32	`json:"stock_after"`
}

type ProductionServiceInterface interface {
	Add(ctx context.Context, vo production.UpsertProductionDocumentVO) error
}

type ProductProviderInterface interface {
	GetById(ctx context.Context, id int) (*entity.Product, error)
}

type ProductionHandler struct {
	service ProductionServiceInterface
}

func New(s ProductionServiceInterface) *ProductionHandler {
	return &ProductionHandler{s}
}

func (ch *ProductionHandler) Create(c *gin.Context) {
	var request UpsertProduction

	if err  := c.BindJSON(&request); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
 
	err := ch.service.Add(c, request.toVO())
	if err != nil {
		httpCode := http.StatusInternalServerError

		if errors.Is(err, production.ErrInsufficientRawMaterialAmount) {
			httpCode = http.StatusUnprocessableEntity
		}
		
		c.Status(httpCode)
		return
	}

	c.Status(http.StatusCreated)
}
