package outgoing

import (
	"context"
	"net/http"
	"vkr/internal/entity/document/outgoing"

	"github.com/gin-gonic/gin"
)

type OutgoingItemDTO struct {
	ProductID int     `json:"product_id" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Price     float64 `json:"price" binding:"required,gte=0"`
}

type UpsertOutgoingDTO struct {
	WarehouseID    int               `json:"warehouse_id" binding:"required"`
	CounterpartyID int               `json:"counterparty_id" binding:"required"`
	Items          []OutgoingItemDTO `json:"items" binding:"required,min=1"`
}

type OutgoingServiceInterface interface {
	Add(ctx context.Context, req outgoing.UpsertOutgoingDocumentVO) error
}

type OutgoingHandler struct {
	service OutgoingServiceInterface
}

func New(isi OutgoingServiceInterface) *OutgoingHandler {
	return &OutgoingHandler{service: isi}
}

func (h *OutgoingHandler) Create(c *gin.Context) {
	var req UpsertOutgoingDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items := make([]outgoing.UpsertOutgoingDocumentItemVO, len(req.Items))
	for i, item := range req.Items {
		items[i] = outgoing.UpsertOutgoingDocumentItemVO{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	err := h.service.Add(c, outgoing.UpsertOutgoingDocumentVO{
		WarehouseID:    req.WarehouseID,
		CounterPartyID: req.CounterpartyID,
		Items:          items,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}