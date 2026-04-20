package counterparty

import (
	"context"
	"errors"
	"net/http"
	"vkr/internal/entity"

	"github.com/gin-gonic/gin"
)

type Counterparty struct {
	ID   int		 `json:"id"`
	Name string      `json:"name" binding:"required"`
	Role string      `json:"role" binding:"required"`
}

type CounterpartyServiceInterface interface {
	Add(ctx context.Context, name, role string) (*entity.Counterparty, error)
	GetById(ctx context.Context, id int) (*entity.Counterparty, error)
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

func (ch *CounterpartyHandler) Get(c *gin.Context) {
    var uriParams struct {
        ID int `uri:"id"`
    } 
    
    if err := c.ShouldBindUri(&uriParams); err != nil {
        c.Status(http.StatusBadRequest)
        return
    }

	item, err := ch.service.GetById(c, uriParams.ID)

	if err != nil {
		if errors.Is(err, entity.ErrCounterpartyNotFound) {
			c.Status(http.StatusNotFound)
			return			
		}

		c.Status(http.StatusInternalServerError)
		return
	}
	
	dto := Counterparty{
		ID: item.ID,
		Name: item.Name,
		Role: item.Role,
	}

	c.IndentedJSON(http.StatusOK, dto)
}

func (ch *CounterpartyHandler) List(c *gin.Context) {
	items, err := ch.service.GetAll(c)
	
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	dtoCollection := make([]Counterparty,0,len(items))

	for _, item := range items {
		dtoCollection = append(dtoCollection, Counterparty{
			ID: item.ID,
			Name: item.Name,
			Role: item.Role,
		})
	}

	c.IndentedJSON(http.StatusOK, dtoCollection)
}

func (ch *CounterpartyHandler) Update(c *gin.Context) {
    var uriParams struct {
        ID int `uri:"id"`
    } 
    
    if err := c.ShouldBindUri(&uriParams); err != nil {
        c.Status(http.StatusBadRequest)
        return
    }

	var counterparty Counterparty

	if err  := c.BindJSON(&counterparty); err != nil {
		return
	}

	item, err := ch.service.Update(
		c, 
		uriParams.ID,
		counterparty.Name,
		counterparty.Role,
	)

	if err != nil {
		if errors.Is(err, entity.ErrCounterpartyNotFound) {
			c.Status(http.StatusNotFound)
			return
		}

		if errors.Is(err, entity.ErrInvalidCounterpartyRole) ||
		   errors.Is(err, entity.ErrInvalidCounterpartyName) {
			c.Status(http.StatusBadRequest)
			return	
		}

		c.Status(http.StatusInternalServerError)
		return
	}

	dto := Counterparty{
		ID: item.ID,
		Name: item.Name,
		Role: item.Role,
	}

	c.IndentedJSON(http.StatusOK, dto)
}

func (ch *CounterpartyHandler) Delete(c *gin.Context) {
    var uriParams struct {
        ID int `uri:"id"`
    } 
    
    if err := c.ShouldBindUri(&uriParams); err != nil {
        c.Status(http.StatusBadRequest)
        return
    }

	err := ch.service.Delete(c, uriParams.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (ch *CounterpartyHandler) Create(c *gin.Context) {
	var counterparty Counterparty

	if err  := c.BindJSON(&counterparty); err != nil {
		return
	}

	item, err := ch.service.Add(
		c, 
		counterparty.Name,
		counterparty.Role,
	)

	if err != nil {
		if errors.Is(err, entity.ErrInvalidCounterpartyRole) ||
		   errors.Is(err, entity.ErrInvalidCounterpartyName) {
			c.Status(http.StatusBadRequest)
			return	
		}

		c.Status(http.StatusInternalServerError)
		return
	}

	dto := Counterparty{
		ID: item.ID,
		Name: item.Name,
		Role: item.Role,
	}

	c.IndentedJSON(http.StatusOK, dto)
}