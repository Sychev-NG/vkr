package repository

import (
	"vkr/internal/entity"
)

type ProductRepository interface {
	GetById(id int) Product
}