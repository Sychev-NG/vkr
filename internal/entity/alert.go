package entity

import (
	"errors"
	"time"
)

var (
	ErrAlertNotFound = errors.New("alert not found")
)

type Alert struct {
	ID          int
	ProductID   int
	WarehouseID int
	Message     string
	CreatedAt   time.Time
	IsResolved  bool
	ResolvedAt  *time.Time
}

type UpsertAlertVO struct {
	ProductID   int
	WarehouseID int
	Message     string
}

type AlertFilter struct {
	ProductID   *int
	WarehouseID *int
	IsResolved  *bool
}