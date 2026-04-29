package entity

import "errors"

var (
	ErrCounterpartyNotFound       = errors.New("counterparty not found")
	ErrCounterpartyDuplicateFound = errors.New("counterparty duplicate found")
	ErrInvalidCounterpartyName    = errors.New("invalid counterparty name")
	ErrInvalidCounterpartyRole    = errors.New("invalid counterparty role")
)

type CounterpartyRole string

const (
	Supplier CounterpartyRole = "supplier"
	Buyer    CounterpartyRole = "buyer"
)

type Counterparty struct {
	ID   int
	Name string
	Role string
}

type CounterpartyFilter struct {
	Role *CounterpartyRole
}