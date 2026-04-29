package entity

import "errors"

var (
	ErrProductNotFound        = errors.New("product not found")
	ErrInvalidProductName     = errors.New("invalid product name")
	ErrInvalidProductUnit     = errors.New("invalid product unit")
	ErrProductDuplicateFound  = errors.New("product duplicate found")
)

type Product struct {
	ID        int
	Name      string
	Unit      string   // kg, piece, pack, liter
	MinStock  float64  // порог для алертов
}
