package document

type DocumentType string

const (
	Incoming 	DocumentType 	= "incoming"
	Outgoing 	DocumentType 	= "outgoing"
	Production 	DocumentType 	= "production"
)

type Document struct {
	DocumentID 	int
	Type		DocumentType
}