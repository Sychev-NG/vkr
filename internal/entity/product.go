package entity

import "errors"

var (
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidProductType = errors.New("invalid product type")
	ErrInvalidProductUnit = errors.New("invalid product unit")
)

type ProductType string

const (
	Raw      ProductType = "raw"
	Finished ProductType = "finished"
)

type ProductUnit string

const (
	KG      ProductUnit = "kg"
)

type Product struct {
	ID       int
	Name     string
	Unit     string
	TypeName string
}