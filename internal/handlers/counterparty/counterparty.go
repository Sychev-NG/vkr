package counterparty

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"vkr/internal/entity"

	"github.com/gin-gonic/gin"
)

type CounterpartyDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name" binding:"required"`
	Role string `json:"role" binding:"required"`
}

type CounterpartyServiceInterface interface {
	Add(ctx context.Context, name, role string) (*entity.Counterparty, error)
	GetByID(ctx context.Context, id int) (*entity.Counterparty, error)
	GetAll(ctx context.Context) ([]entity.Counterparty, error)
	Update(ctx context.Context, id int, name, role string) (*entity.Counterparty, error)
	Delete(ctx context.Context, id int) error
}

type CounterpartyHandler struct {
	service CounterpartyServiceInterface
}

func New(csi CounterpartyServiceInterface) *CounterpartyHandler {
	return &CounterpartyHandler{csi}
}

func (h *CounterpartyHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	item, err := h.service.GetByID(c, id)
	if err != nil {
		if errors.Is(err, entity.ErrCounterpartyNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, CounterpartyDTO{
		ID:   item.ID,
		Name: item.Name,
		Role: item.Role,
	})
}

func (h *CounterpartyHandler) List(c *gin.Context) {
	items, err := h.service.GetAll(c)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	dtos := make([]CounterpartyDTO, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, CounterpartyDTO{
			ID:   item.ID,
			Name: item.Name,
			Role: item.Role,
		})
	}
	c.JSON(http.StatusOK, dtos)
}

func (h *CounterpartyHandler) Create(c *gin.Context) {
	var dto CounterpartyDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.Add(c, dto.Name, dto.Role)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidCounterpartyName) ||
			errors.Is(err, entity.ErrInvalidCounterpartyRole) {
			c.Status(http.StatusBadRequest)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, CounterpartyDTO{
		ID:   item.ID,
		Name: item.Name,
		Role: item.Role,
	})
}

func (h *CounterpartyHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	var dto CounterpartyDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.Update(c, id, dto.Name, dto.Role)
	if err != nil {
		if errors.Is(err, entity.ErrCounterpartyNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if errors.Is(err, entity.ErrInvalidCounterpartyName) ||
			errors.Is(err, entity.ErrInvalidCounterpartyRole) {
			c.Status(http.StatusBadRequest)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, CounterpartyDTO{
		ID:   item.ID,
		Name: item.Name,
		Role: item.Role,
	})
}

func (h *CounterpartyHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	err = h.service.Delete(c, id)
	if err != nil {
		if errors.Is(err, entity.ErrCounterpartyNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}