package document

type DocumentStatus string

const (
	Draft 		DocumentStatus 	= "draft"
	Confirmed 	DocumentStatus 	= "confirmed"
	Posted 		DocumentStatus 	= "posted"
	Cancelled 	DocumentStatus 	= "cancelled"
	Canceled 	DocumentStatus 	= "canceled"
	Archived 	DocumentStatus 	= "archived"
)