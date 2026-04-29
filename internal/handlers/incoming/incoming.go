package incoming

import (
	"context"
	"net/http"
	"vkr/internal/entity/document/incoming"

	"github.com/gin-gonic/gin"
)

type IncomingItemDTO struct {
	ProductID int     `json:"product_id" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Price     float64 `json:"price" binding:"required,gte=0"`
}

type UpsertIncomingDTO struct {
	WarehouseID    int               `json:"warehouse_id" binding:"required"`
	CounterpartyID int               `json:"counterparty_id" binding:"required"`
	Items          []IncomingItemDTO `json:"items" binding:"required,min=1"`
}

type IncomingServiceInterface interface {
	Add(ctx context.Context, req incoming.UpsertIncomingDocumentVO) error
}

type IncomingHandler struct {
	service IncomingServiceInterface
}

func New(isi IncomingServiceInterface) *IncomingHandler {
	return &IncomingHandler{service: isi}
}

func (h *IncomingHandler) Create(c *gin.Context) {
	var req UpsertIncomingDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items := make([]incoming.UpsertIncomingDocumentItemVO, len(req.Items))
	for i, item := range req.Items {
		items[i] = incoming.UpsertIncomingDocumentItemVO{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	err := h.service.Add(c, incoming.UpsertIncomingDocumentVO{
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