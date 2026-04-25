package production

import (
	"errors"
	"time"

	"vkr/internal/entity/document"
)

var (
	ErrProductionDocumentNotFound 	    = errors.New("production document not found")
	ErrInsufficientRawMaterialAmount 	= errors.New("insufficient raw material amount")
)

type ProductionDocument struct {
	ID       		int
	WarehouseID		int
	Date			time.Time
	Items 			[]ProductionDocumentItem
}

func (doc *ProductionDocument) ToDocument() document.Document {
	return document.Document{DocumentID: doc.ID, Type: document.Production}
}

type ProductionDocumentItem struct {
	ID						int
	FinishedMaterialID		int
	RecipeID		        int
	Quantity				float32
}

type UpsertProductionDocumentVO struct {
	WarehouseID		int
	Items 			[]UpsertProductionDocumentItemVO
}

type UpsertProductionDocumentItemVO struct {
	FinishedMaterialID		int
	RecipeID		        int
	Quantity				float32
}