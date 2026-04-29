package entity

import "errors"

var (
	ErrAssemblyNotFound       = errors.New("assembly not found")
	ErrAssemblyDuplicateFound = errors.New("assembly duplicate found")
	ErrInvalidAssemblyName    = errors.New("invalid assembly name")
	ErrInvalidComponent       = errors.New("invalid component")
	ErrInsufficientComponents = errors.New("insufficient components for assembly")
)

type Assembly struct {
	ID               int
	Name             string
	OutputProductID  int
	OutputQuantity   float64  // сколько получается из спецификации (обычно 1)
	Components       []AssemblyComponent
}

type AssemblyComponent struct {
	ID         int
	AssemblyID int
	ProductID  int
	Quantity   float64  // сколько нужно компонентов на OutputQuantity
}

type UpsertAssemblyVO struct {
	Name            string
	OutputProductID int
	OutputQuantity  float64
	Components      []UpsertComponentVO
}

type UpsertComponentVO struct {
	ProductID int
	Quantity  float64
}

type AssemblyRequirement struct {
	ProductID     int
	ProductName   string
	Required      float64
	Available     float64
	Sufficient    bool
	TotalCost     float64
}