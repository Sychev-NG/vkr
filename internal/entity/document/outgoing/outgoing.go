package outgoing

import (
	"errors"
	"time"

	"vkr/internal/entity/document"
)

var (
	ErrOutgoingDocumentNotFound 	= errors.New("outgoing document not found")
	ErrBuyerNotFound 				= errors.New("buyer not found")
	ErrInvalidBuyer 				= errors.New("invalid counterparty")
)

type OutgoingDocument struct {
	ID       		int
	WarehouseID		int
	CounterPartyID	int
	Date			time.Time
	Items 			[]OutgoingDocumentItem
}

func (doc *OutgoingDocument) ToDocument() document.Document {
	return document.Document{DocumentID: doc.ID, Type: document.Outgoing}
}

type OutgoingDocumentItem struct {
	ID						int
	FinishedMaterialID		int
	Quantity				float32
	Price					float32
}

type UpsertOutgoingDocumentVO struct {
	WarehouseID		int
	CounterPartyID	int
	Items 			[]UpsertOutgoingDocumentItemVO
}

type UpsertOutgoingDocumentItemVO struct {
	FinishedMaterialID		int
	Quantity				float32
	Price					float32
}