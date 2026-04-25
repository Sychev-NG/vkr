package production

import (
	"errors"
	"time"

	"vkr/internal/entity/document"
)

var (
	ErrProductionDocumentNotFound 	    = errors.New("production document not found")
)

type ProductionDocument struct {
	ID       		int
	WarehouseID		int
	RecipeID		int
	Quantity		float32
	Date			time.Time
}

func (doc *ProductionDocument) ToDocument() document.Document {
	return document.Document{DocumentID: doc.ID, Type: document.Production}
}

type UpsertProductionDocumentVO struct {
	WarehouseID		int
	RecipeID		int
	Quantity		float32
}