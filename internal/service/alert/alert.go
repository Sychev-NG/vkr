package alert

import (
	"context"
	"log"
	"vkr/internal/entity"
)

type AlertSaver interface{
	Resolve(ctx context.Context, id int) error
	Create(ctx context.Context, vo entity.UpsertAlertVO) (*entity.Alert, error)
}

type AlertProvider interface{
	GetByFilter(ctx context.Context, filter entity.AlertFilter) ([]entity.Alert, error)
}

type AlertService struct {
	s AlertSaver
	p AlertProvider
}

func New(s AlertSaver, p AlertProvider) *AlertService {
	return &AlertService{s: s, p: p}
}

func (s *AlertService) GetAlerts(ctx context.Context, filter entity.AlertFilter) ([]entity.Alert, error) {
	alerts, err := s.p.GetByFilter(ctx, filter)
	if err != nil {
		log.Printf("AlertService::GetAlerts Error - %v", err)
		return nil, err
	}
	return alerts, nil
}

func (s *AlertService) Add(ctx context.Context, vo entity.UpsertAlertVO) (*entity.Alert, error) {
	return s.s.Create(ctx, vo)
}

func (s *AlertService) Resolve(ctx context.Context, id int) error {
	err := s.s.Resolve(ctx, id)
	if err != nil {
		log.Printf("AlertService::Resolve Error - %v", err)
		return err
	}
	return nil
}