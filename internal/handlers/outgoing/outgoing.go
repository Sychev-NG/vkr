package outgoing

import (
	"context"
	"errors"
	"net/http"
	"vkr/internal/entity"
	"vkr/internal/entity/document/outgoing"

	"github.com/gin-gonic/gin"
)

type UpsertOutgoing struct {
	CounterpartyID	int 					`json:"counterparty_id" binding:"required"`
	WarehouseID		int 					`json:"warehouse_id" binding:"required"`
	Items 			[]UpsertOutgoingItem 	`json:"items" binding:"required"`
}

type UpsertOutgoingItem struct {
	FinishedMaterialID 	int			`json:"product_id" binding:"required"`
	Quantity 			float32		`json:"quantity" binding:"required"`
	Price 				float32 	`json:"price" binding:"required"`
}

func (r *UpsertOutgoing) toVO() outgoing.UpsertOutgoingDocumentVO {
	var items []outgoing.UpsertOutgoingDocumentItemVO

	for _, item := range r.Items {
		items = append(items, outgoing.UpsertOutgoingDocumentItemVO{
			FinishedMaterialID: item.FinishedMaterialID,
			Quantity: item.Quantity,
			Price: item.Price,
		})
	}

	return outgoing.UpsertOutgoingDocumentVO{
		CounterPartyID: r.CounterpartyID,
		WarehouseID: r.WarehouseID,
		Items: items,
	}
}

type Outgoing struct {
	DocumentID	int 			`json:"document_id"`
	Items 		[]OutgoingItem 	`json:"movements"`
}

type OutgoingItem struct {
	ProductID 	int 	`json:"product_id"`
	ProductName	int 	`json:"product_name"`
	Quantity 	float32	`json:"quantity"`
}

type OutgoingServiceInterface interface {
	Add(ctx context.Context, vo outgoing.UpsertOutgoingDocumentVO) error
	GetAll(ctx context.Context) ([]outgoing.OutgoingDocument, error)
}

type ProductProviderInterface interface {
	GetById(ctx context.Context, id int) (*entity.Product, error)
}

type OutgoingHandler struct {
	service OutgoingServiceInterface
	productProvider ProductProviderInterface
}

func New(s OutgoingServiceInterface, pp ProductProviderInterface) *OutgoingHandler {
	return &OutgoingHandler{s, pp}
}

func (ch *OutgoingHandler) Create(c *gin.Context) {
	var request UpsertOutgoing

	if err  := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
 
	err := ch.service.Add(c, request.toVO())
	if err != nil {
		if errors.Is(err, outgoing.ErrBuyerNotFound) ||
			errors.Is(err, entity.ErrInvalidFinishedMaterial) ||
		   errors.Is(err, entity.ErrFinishedProductNotFound) {
			c.Status(http.StatusBadRequest)
		}

		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusCreated)
}
