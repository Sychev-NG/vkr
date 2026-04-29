package assemblyorder

import (
	// "errors"
	"context"
	"net/http"
	"vkr/internal/entity/document/assembly"

	"github.com/gin-gonic/gin"
)

type AssemblyOrderDTO struct {
	AssemblyID      int     `json:"assembly_id" binding:"required"`
	WarehouseID     int     `json:"warehouse_id" binding:"required"`
	QuantityToBuild float64 `json:"quantity" binding:"required"`
}

type AssemblyOrderServiceInterface interface {
	Create(ctx context.Context, req assembly.UpsertAssemblyOrderVO) error
	GetByID(ctx context.Context, id int) (*assembly.AssemblyOrder, error)
	GetAll(ctx context.Context) ([]assembly.AssemblyOrder, error)
}

type AssemblyOrderHandler struct {
	service AssemblyOrderServiceInterface
}

func New(aosi AssemblyOrderServiceInterface) *AssemblyOrderHandler {
	return &AssemblyOrderHandler{service: aosi}
}

func (h *AssemblyOrderHandler) Create(c *gin.Context) {
	var dto AssemblyOrderDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.Create(c, assembly.UpsertAssemblyOrderVO{
		AssemblyID:      dto.AssemblyID,
		WarehouseID:     dto.WarehouseID,
		QuantityToBuild: dto.QuantityToBuild,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

func (h *AssemblyOrderHandler) Get(c *gin.Context) {
	var uriParams struct {
		ID int `uri:"id"`
	}

	if err := c.ShouldBindUri(&uriParams); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	order, err := h.service.GetByID(c, uriParams.ID)
	if err != nil {
		// if errors.Is(err, docAssembly.ErrAssemblyOrderNotFound) {
		// 	c.Status(http.StatusNotFound)
		// 	return
		// }
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *AssemblyOrderHandler) List(c *gin.Context) {
	orders, err := h.service.GetAll(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// func (h *AssemblyOrderHandler) CheckAvailability(c *gin.Context) {
// 	var req struct {
// 		AssemblyID      int     `json:"assembly_id" binding:"required"`
// 		WarehouseID     int     `json:"warehouse_id" binding:"required"`
// 		QuantityToBuild float64 `json:"quantity" binding:"required"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	requirements, err := h.service.CheckAvailability(c, req.AssemblyID, req.WarehouseID, req.QuantityToBuild)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	response := make([]gin.H, len(requirements))
// 	allAvailable := true
// 	for i, req := range requirements {
// 		response[i] = gin.H{
// 			"product_id": req.ProductID,
// 			"required":   req.Required,
// 			"available":  req.Available,
// 			"sufficient": req.Available >= req.Required,
// 		}
// 		if req.Available < req.Required {
// 			allAvailable = false
// 		}
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"components":     response,
// 		"all_available": allAvailable,
// 	})
// }