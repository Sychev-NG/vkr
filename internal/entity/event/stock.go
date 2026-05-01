package event

type StockEvent struct {
	ProductID int
	WarehouseID int
	OldQuantity float64
	NewQuantity float64
}

func (e *StockEvent) EventName() string {
	return "stock.changed"
}