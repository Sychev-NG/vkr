package alert

import (
	"context"
	"net/http"
	"strconv"
	"vkr/internal/entity"

	"github.com/gin-gonic/gin"
)

type AlertDTO struct {
	ID            int     `json:"id"`
	ProductID     int     `json:"product_id"`
	ProductName   string  `json:"product_name"`
	WarehouseID   int     `json:"warehouse_id"`
	WarehouseName string  `json:"warehouse_name"`
	Message       string  `json:"message"`
	CreatedAt     string  `json:"created_at"`
	IsResolved    bool    `json:"is_resolved"`
	ResolvedAt    *string `json:"resolved_at,omitempty"`
}

type AlertServiceInterface interface {
	GetAlerts(ctx context.Context, filter entity.AlertFilter) ([]entity.Alert, error)
	Resolve(ctx context.Context, id int) error
}

type ProductProvider interface {
	GetByID(ctx context.Context, id int) (*entity.Product, error)
}

type WarehouseProvider interface {
	GetByID(ctx context.Context, id int) (*entity.Warehouse, error)
}

type AlertHandler struct {
	service       AlertServiceInterface
	productProvider   ProductProvider
	warehouseProvider WarehouseProvider
}

func New(asi AlertServiceInterface, pr ProductProvider, wr WarehouseProvider) *AlertHandler {
	return &AlertHandler{
		service:       		asi,
		productProvider:   	pr,
		warehouseProvider: 	wr,
	}
}

func (h *AlertHandler) List(c *gin.Context) {
	filter := entity.AlertFilter{}

	if resolvedStr := c.Query("resolved"); resolvedStr != "" {
		resolved := resolvedStr == "true"
		filter.IsResolved = &resolved
	} else {
		falseVal := false
		filter.IsResolved = &falseVal
	}

	if productIDStr := c.Query("product_id"); productIDStr != "" {
		productID, err := strconv.Atoi(productIDStr)
		if err == nil {
			filter.ProductID = &productID
		}
	}

	if warehouseIDStr := c.Query("warehouse_id"); warehouseIDStr != "" {
		warehouseID, err := strconv.Atoi(warehouseIDStr)
		if err == nil {
			filter.WarehouseID = &warehouseID
		}
	}

	alerts, err := h.service.GetAlerts(c, filter)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	dtos := make([]AlertDTO, 0, len(alerts))
	for _, alert := range alerts {
		product, err := h.productProvider.GetByID(c, alert.ProductID)
		if err != nil {
			continue
		}
		warehouse, err := h.warehouseProvider.GetByID(c, alert.WarehouseID)
		if err != nil {
			continue
		}

		dto := AlertDTO{
			ID:            alert.ID,
			ProductID:     alert.ProductID,
			ProductName:   product.Name,
			WarehouseID:   alert.WarehouseID,
			WarehouseName: warehouse.Name,
			Message:       alert.Message,
			CreatedAt:     alert.CreatedAt.Format("2006-01-02T15:04:05Z"),
			IsResolved:    alert.IsResolved,
		}
		if alert.ResolvedAt != nil {
			resolvedAt := alert.ResolvedAt.Format("2006-01-02T15:04:05Z")
			dto.ResolvedAt = &resolvedAt
		}
		dtos = append(dtos, dto)
	}

	c.JSON(http.StatusOK, dtos)
}

func (h *AlertHandler) Resolve(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	err = h.service.Resolve(c, id)
	if err != nil {
		if err == entity.ErrAlertNotFound {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}