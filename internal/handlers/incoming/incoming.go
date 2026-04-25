package incoming

import (
	"context"
	"errors"
	"net/http"
	"vkr/internal/entity"
	"vkr/internal/entity/document/incoming"

	"github.com/gin-gonic/gin"
)

type UpsertIncoming struct {
	CounterpartyID	int 					`json:"counterparty_id" binding:"required"`
	WarehouseID		int 					`json:"warehouse_id" binding:"required"`
	Items 			[]UpsertIncomingItem 	`json:"items" binding:"required"`
}

type UpsertIncomingItem struct {
	RawMaterialID 	int 		`json:"product_id" binding:"required"`
	Quantity 		float32		`json:"quantity" binding:"required"`
	Price 			float32 	`json:"price" binding:"required"`
}

func (r *UpsertIncoming) toVO() incoming.UpsertIncomingDocumentVO {
	var items []incoming.UpsertIncomingDocumentItemVO

	for _, item := range r.Items {
		items = append(items, incoming.UpsertIncomingDocumentItemVO{
			RawMaterialID: item.RawMaterialID,
			Quantity: item.Quantity,
			Price: item.Price,
		})
	}

	return incoming.UpsertIncomingDocumentVO{
		CounterPartyID: r.CounterpartyID,
		WarehouseID: r.WarehouseID,
		Items: items,
	}
}

type Incoming struct {
	DocumentID	int 			`json:"document_id"`
	Status 		string 			`json:"status"`
	Items 		[]IncomingItem 	`json:"movements"`
}

type IncomingItem struct {
	ProductID 	int 	`json:"product_id"`
	Quantity 	float32	`json:"quantity"`
	StockBefore float32	`json:"stock_before"`
	StockAfter 	float32	`json:"stock_after"`
}

type IncomingServiceInterface interface {
	Add(ctx context.Context, vo incoming.UpsertIncomingDocumentVO) error
	GetAll(ctx context.Context) ([]incoming.IncomingDocument, error)
}

type ProductProviderInterface interface {
	GetById(ctx context.Context, id int) (*entity.Product, error)
}

type IncomingHandler struct {
	service IncomingServiceInterface
	productProvider ProductProviderInterface
}

func New(s IncomingServiceInterface, pp ProductProviderInterface) *IncomingHandler {
	return &IncomingHandler{s, pp}
}

func (ch *IncomingHandler) Create(c *gin.Context) {
	var request UpsertIncoming

	if err  := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
 
	err := ch.service.Add(c, request.toVO())
	if err != nil {
		if errors.Is(err, incoming.ErrSupplierNotFound) ||
		   errors.Is(err, entity.ErrInvalidRawMaterial) ||
		   errors.Is(err, entity.ErrRawProductNotFound) {
			c.Status(http.StatusBadRequest)
		}

		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
