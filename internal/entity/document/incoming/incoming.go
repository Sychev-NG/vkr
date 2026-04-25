package incoming

import (
	"errors"
	"time"

	"vkr/internal/entity/document"
)

var (
	ErrIncomingDocumentNotFound 	= errors.New("incoming document not found")
	ErrSupplierNotFound 			= errors.New("supplier not found")
	ErrInvalidSupplier 				= errors.New("invalid supplier")
)

type IncomingDocument struct {
	ID       		int
	WarehouseID		int
	CounterPartyID	int
	Date			time.Time
	Items 			[]IncomingDocumentItem
}

func (doc *IncomingDocument) ToDocument() document.Document {
	return document.Document{DocumentID: doc.ID, Type: document.Incoming}
}

type IncomingDocumentItem struct {
	ID						int
	RawMaterialID			int
	Quantity				float32
	Price					float32
}

type UpsertIncomingDocumentVO struct {
	WarehouseID		int
	CounterPartyID	int
	Items 			[]UpsertIncomingDocumentItemVO
}

type UpsertIncomingDocumentItemVO struct {
	RawMaterialID			int
	Quantity				float32
	Price					float32
}