package entity

import "errors"

var (
	ErrProductNotFound = errors.New("product not found")
)

type Product struct {
	ID       int
	Name     string
	Unit     string
	TypeName string
}