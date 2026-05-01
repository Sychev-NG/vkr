package subscriber

import (
	"context"
	"log"
	"vkr/internal/entity"
	domainEvent "vkr/internal/entity/event"
	"vkr/internal/event"
)

type AlertService interface {
	Add(ctx context.Context, vo entity.UpsertAlertVO) (*entity.Alert, error)
}

type ProductProvider interface {
	GetByID(ctx context.Context, id int) (*entity.Product, error)
}

type StockAlertSubscriber struct {
	as	AlertService
	pp	ProductProvider
}

func New(as AlertService, pp ProductProvider) *StockAlertSubscriber {
	return &StockAlertSubscriber{as: as, pp: pp}
}

func (s *StockAlertSubscriber) Handle(ctx context.Context, event event.Event) error {
	stockAlert, ok := event.(*domainEvent.StockEvent)
	if !ok {
		return nil
	}

	product, err := s.pp.GetByID(ctx, stockAlert.ProductID)
	if err != nil {
		log.Printf("StockAlertSubscriber::Handle pp.GetByID Error - %v", err)
		return err
	}

	if stockAlert.NewQuantity <= product.MinStock {
		_, err := s.as.Add(ctx, entity.UpsertAlertVO{
			ProductID: stockAlert.ProductID,
			WarehouseID: stockAlert.WarehouseID,
			Message: "Низкий остаток!",
		})

		if err != nil {
			log.Printf("StockAlertSubscriber::Handle as.Add Error - %v", err)
			return err
		}
	}

	return nil
}