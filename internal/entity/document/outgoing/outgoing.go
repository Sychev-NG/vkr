package outgoing

import (
	"errors"
	"time"

	"vkr/internal/entity/document"
)

var (
	ErrOutgoingDocumentNotFound = errors.New("outgoing document not found")
	ErrSupplierNotFound         = errors.New("supplier not found")
)

type OutgoingDocument struct {
	ID             int
	WarehouseID    int
	CounterPartyID int
	Date           time.Time
	Items          []OutgoingDocumentItem
}

func (doc *OutgoingDocument) ToDocument() document.Document {
	return document.Document{DocumentID: doc.ID, Type: document.Outgoing}
}

type OutgoingDocumentItem struct {
	ID        int
	ProductID int
	Quantity  float64
	Price     float64
	UnitCost  float64
}

type UpsertOutgoingDocumentVO struct {
	WarehouseID    int
	CounterPartyID int
	Items          []UpsertOutgoingDocumentItemVO
}

type UpsertOutgoingDocumentItemVO struct {
	ProductID int
	Quantity  float64
	Price     float64
}