package entity

import (
	"errors"
	"time"
)

var (
	ErrBatchNotFound      = errors.New("batch not found")
	ErrInsufficientBatch  = errors.New("insufficient quantity in batch")
)

type BatchType string

const (
	BatchTypeIncoming  BatchType = "incoming"
	BatchTypeAssembly  BatchType = "assembly"
)

type Batch struct {
	ID                	int
	ProductID         	int
	WarehouseID       	int
	DocumentType      	string
	DocumentID	      	int
	QuantityRemaining 	float64
	UnitCost 			float64
	CreatedAt         	time.Time
}

type BatchMovement struct {
	ID                     int
	BatchID                int
	DocumentType           string
	DocumentID             int
	Quantity               float64
	CreatedAt              time.Time
}

type UpsertBatchVO struct {
	ProductID         	int
	WarehouseID       	int
	DocumentID  	  	int
	DocumentType      	string
	QuantityRemaining 	float64
	UnitCost 			float64
}

type BatchFilter struct {
	ProductID   int
	WarehouseID int
}

type Movement struct {
	ID                	int    
	BatchID             int    
	ProductID         	int    
	WarehouseID       	int    
	StockMovement		float64
	UnitCost			float64
	Date         		time.Time 
}

type MovementFilter struct {
	DocumentID     	int
	DocumentType   	string
	BatchID     	int
	ProductID   	int
	WarehouseID 	int
}

// type BatchOrderBy string

// const (
// 	BatchOrderByCreatedAt  BatchOrderBy = "created_at"
// )

// type BatchOrderDirection string

// const (
// 	BatchOrderDirectionDesc  BatchOrderBy = "desc"
// 	BatchOrderDirectionAsc  BatchOrderBy = "asc"
// )

// type BatchOrder struct {
// 	By   		BatchOrderBy
// 	Direction 	BatchOrderDirection
// }