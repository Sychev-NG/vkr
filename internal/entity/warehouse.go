package entity

import "errors"

var (
	ErrWarehouseNotFound       = errors.New("warehouse not found")
	ErrWarehouseDuplicateFound = errors.New("warehouse duplicate found")
	ErrInvalidWarehouseName    = errors.New("invalid warehouse name")
	ErrInvalidWarehouseAddress = errors.New("invalid warehouse address")
)

type Warehouse struct {
	ID      int
	Name    string
	Address string
}

type WarehouseFilter struct {
	ID   *int
	Name *string
}