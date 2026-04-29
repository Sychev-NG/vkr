package warehouse

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"vkr/internal/entity"

	"github.com/gin-gonic/gin"
)

type WarehouseDTO struct {
	ID      int    `json:"id"`
	Name    string `json:"name" binding:"required"`
	Address string `json:"address" binding:"required"`
}

type WarehouseServiceInterface interface {
	Add(ctx context.Context, name, address string) (*entity.Warehouse, error)
	GetByID(ctx context.Context, id int) (*entity.Warehouse, error)
	GetAll(ctx context.Context) ([]entity.Warehouse, error)
	Update(ctx context.Context, id int, name, address string) (*entity.Warehouse, error)
	Delete(ctx context.Context, id int) error
}

type WarehouseHandler struct {
	service WarehouseServiceInterface
}

func New(wsi WarehouseServiceInterface) *WarehouseHandler {
	return &WarehouseHandler{wsi}
}

func (h *WarehouseHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	item, err := h.service.GetByID(c, id)
	if err != nil {
		if errors.Is(err, entity.ErrWarehouseNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, WarehouseDTO{
		ID:      item.ID,
		Name:    item.Name,
		Address: item.Address,
	})
}

func (h *WarehouseHandler) List(c *gin.Context) {
	items, err := h.service.GetAll(c)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	dtos := make([]WarehouseDTO, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, WarehouseDTO{
			ID:      item.ID,
			Name:    item.Name,
			Address: item.Address,
		})
	}
	c.JSON(http.StatusOK, dtos)
}

func (h *WarehouseHandler) Create(c *gin.Context) {
	var dto WarehouseDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.Add(c, dto.Name, dto.Address)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidWarehouseName) {
			c.Status(http.StatusBadRequest)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, WarehouseDTO{
		ID:      item.ID,
		Name:    item.Name,
		Address: item.Address,
	})
}

func (h *WarehouseHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	var dto WarehouseDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.Update(c, id, dto.Name, dto.Address)
	if err != nil {
		if errors.Is(err, entity.ErrWarehouseNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if errors.Is(err, entity.ErrInvalidWarehouseName) {
			c.Status(http.StatusBadRequest)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, WarehouseDTO{
		ID:      item.ID,
		Name:    item.Name,
		Address: item.Address,
	})
}

func (h *WarehouseHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	err = h.service.Delete(c, id)
	if err != nil {
		if errors.Is(err, entity.ErrWarehouseNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}