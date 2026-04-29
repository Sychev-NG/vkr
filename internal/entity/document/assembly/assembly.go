package assembly

import (
	"errors"
	"time"

	"vkr/internal/entity/document"
)

var (
	ErrAssemblyOrderNotFound = errors.New("assembly order not found")
)

type AssemblyOrder struct {
	ID              int
	AssemblyID      int
	WarehouseID     int
	QuantityToBuild float64
	Date            time.Time
	Consumptions    []AssemblyOrderConsumption
	Outputs         []AssemblyOrderOutput
}

func (doc *AssemblyOrder) ToDocument() document.Document {
	return document.Document{DocumentID: doc.ID, Type: document.Assembly}
}

type AssemblyOrderConsumption struct {
	ID        int
	OrderID   int
	ProductID int
	Quantity  float64
	UnitCost  float64
	TotalCost float64
}

type AssemblyOrderOutput struct {
	ID        int
	OrderID   int
	ProductID int
	Quantity  float64
	UnitCost  float64
	TotalCost float64
}

type UpsertAssemblyOrderVO struct {
	AssemblyID      int
	WarehouseID     int
	QuantityToBuild float64
}