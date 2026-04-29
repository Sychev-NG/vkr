package document

import "errors"

type DocumentType string

var (
	ErrUnknownDocumentType = errors.New("assembly not found")
)

const (
	Incoming DocumentType = "incoming"
	Outgoing DocumentType = "outgoing"
	Assembly DocumentType = "assembly" // вместо production
)

type Document struct {
	DocumentID int
	Type       DocumentType
}