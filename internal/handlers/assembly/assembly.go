package assembly

import (
	"context"
	"errors"
	"net/http"
	"vkr/internal/entity"

	"github.com/gin-gonic/gin"
)

type AssemblyDTO struct {
	ID              int                 `json:"id"`
	Name            string              `json:"name" binding:"required"`
	OutputProductID int                 `json:"output_product_id" binding:"required"`
	OutputQuantity  float64             `json:"output_quantity" binding:"required"`
	Components      []ComponentDTO      `json:"components" binding:"required"`
}

type ComponentDTO struct {
	ProductID int     `json:"product_id" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required"`
}

type AssemblyServiceInterface interface {
	Add(ctx context.Context, vo entity.UpsertAssemblyVO) error
	GetByID(ctx context.Context, id int) (*entity.Assembly, error)
	GetAll(ctx context.Context) ([]entity.Assembly, error)
	Update(ctx context.Context, id int, vo entity.UpsertAssemblyVO) error
	Delete(ctx context.Context, id int) error
}

type AssemblyHandler struct {
	service AssemblyServiceInterface
}

func New(service AssemblyServiceInterface) *AssemblyHandler {
	return &AssemblyHandler{service: service}
}

func (ah *AssemblyHandler) Get(c *gin.Context) {
	var uriParams struct {
		ID int `uri:"id"`
	}

	if err := c.ShouldBindUri(&uriParams); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	assembly, err := ah.service.GetByID(c, uriParams.ID)
	if err != nil {
		if errors.Is(err, entity.ErrAssemblyNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	dto := toDTO(assembly)
	c.JSON(http.StatusOK, dto)
}

func (ah *AssemblyHandler) List(c *gin.Context) {
	assemblies, err := ah.service.GetAll(c)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	dtos := make([]AssemblyDTO, 0, len(assemblies))
	for _, assembly := range assemblies {
		dtos = append(dtos, toDTO(&assembly))
	}

	c.JSON(http.StatusOK, dtos)
}

func (ah *AssemblyHandler) Update(c *gin.Context) {
	var uriParams struct {
		ID int `uri:"id"`
	}

	if err := c.ShouldBindUri(&uriParams); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	var dto AssemblyDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vo := toVO(dto)
	err := ah.service.Update(c, uriParams.ID, vo)
	if err != nil {
		if errors.Is(err, entity.ErrAssemblyNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if errors.Is(err, entity.ErrProductNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "product not found"})
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	updated, err := ah.service.GetByID(c, uriParams.ID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, toDTO(updated))
}

func (ah *AssemblyHandler) Delete(c *gin.Context) {
	var uriParams struct {
		ID int `uri:"id"`
	}

	if err := c.ShouldBindUri(&uriParams); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	err := ah.service.Delete(c, uriParams.ID)
	if err != nil {
		if errors.Is(err, entity.ErrAssemblyNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (ah *AssemblyHandler) Create(c *gin.Context) {
	var dto AssemblyDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vo := toVO(dto)
	err := ah.service.Add(c, vo)
	if err != nil {
		if errors.Is(err, entity.ErrProductNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "product not found"})
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusCreated)
}

func toDTO(assembly *entity.Assembly) AssemblyDTO {
	components := make([]ComponentDTO, len(assembly.Components))
	for i, comp := range assembly.Components {
		components[i] = ComponentDTO{
			ProductID: comp.ProductID,
			Quantity:  comp.Quantity,
		}
	}

	return AssemblyDTO{
		ID:              assembly.ID,
		Name:            assembly.Name,
		OutputProductID: assembly.OutputProductID,
		OutputQuantity:  assembly.OutputQuantity,
		Components:      components,
	}
}

func toVO(dto AssemblyDTO) entity.UpsertAssemblyVO {
	components := make([]entity.UpsertComponentVO, len(dto.Components))
	for i, comp := range dto.Components {
		components[i] = entity.UpsertComponentVO{
			ProductID: comp.ProductID,
			Quantity:  comp.Quantity,
		}
	}

	return entity.UpsertAssemblyVO{
		Name:            dto.Name,
		OutputProductID: dto.OutputProductID,
		OutputQuantity:  dto.OutputQuantity,
		Components:      components,
	}
}