package entity

import (
	"errors"
	"time"
)

var (
	ErrMovementNotFound 		= errors.New("movement not found")
)

type DocumentType string 

const (
	Incoming 	DocumentType = "incoming"
	Outgoing 	DocumentType = "outgoing"
	Production 	DocumentType = "production"
)

type Movement struct {
	ID       		int
	ProductID		int
	WarehouseID		int
	DocumentID		int
	DocumentType	DocumentType
	Quantity		float32
	Date			time.Time
}

type MovementFilter struct {
    ProductID    int
    WarehouseID  int
}