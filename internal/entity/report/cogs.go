package report

import "time"

type COGSDocument struct {
	DocumentID       int
	Date             time.Time
	CounterpartyID   int
	CounterpartyName string
	Revenue          float64 // Общая прибыль
	Cogs             float64 // Себестоимость
	Profit           float64 // Валовая прибыль
	Items            []COGSItem
}

type COGSItem struct {
	ProductID    int
	ProductName  string
	ProductUnit  string
	Quantity     float64
	SellingPrice float64
	UnitCost     float64
	Revenue      float64 // Общая прибыль
	Cogs         float64 // Себестоимость
	Profit       float64 // Валовая прибыль
}