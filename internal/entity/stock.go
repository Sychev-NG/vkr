package entity

import (
	"errors"
)

var (
	ErrStockNotFound 		= errors.New("stock not found")
	ErrInvalidQuantity 		= errors.New("invalid quantity")
	ErrInsufficientStock   	= errors.New("insufficient stock")
)

type Stock struct {
	ID       	int
	ProductID	int
	WarehouseID	int
	Quantity	float32
}

type StockFilter struct {
	ProductID	int
	WarehouseID	int
}