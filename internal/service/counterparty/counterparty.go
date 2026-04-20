package counterparty

import (
	"context"
	"strings"
	"vkr/internal/entity"
)

type CounterpartySaver interface {
	Add(ctx context.Context, name, role string) (*entity.Counterparty, error)
	Update(ctx context.Context, id int, name, role string) (*entity.Counterparty, error)
	Delete(ctx context.Context, id int) (error)
}

type CounterpartyProvider interface{
	GetById(ctx context.Context, id int) (*entity.Counterparty, error)
	GetAll(ctx context.Context) ([]entity.Counterparty, error)
}

type CounterpartyService struct {
	saver 		CounterpartySaver
	provider	CounterpartyProvider 
}

func New(ps CounterpartySaver, pp CounterpartyProvider) *CounterpartyService {
	return &CounterpartyService{ps, pp}
}

func (ps *CounterpartyService) Add(ctx context.Context, name, role string) (*entity.Counterparty, error) {
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return nil, entity.ErrInvalidCounterpartyName
	}

	cr := entity.CounterpartyRole(role)
	if cr != entity.Supplier && cr != entity.Buyer {
		return nil, entity.ErrInvalidCounterpartyRole
	}

	return ps.saver.Add(ctx, name, role)	
}

func (ps *CounterpartyService) GetById(ctx context.Context, id int) (*entity.Counterparty, error) {
	return ps.provider.GetById(ctx, id)
}

func (ps *CounterpartyService) GetAll(ctx context.Context) ([]entity.Counterparty, error) {
	return ps.provider.GetAll(ctx)
}

func (ps *CounterpartyService) Update(ctx context.Context, id int, name, role string) (*entity.Counterparty, error) {
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return nil, entity.ErrInvalidCounterpartyName
	}

	cr := entity.CounterpartyRole(role)
	if cr != entity.Supplier && cr != entity.Buyer {
		return nil, entity.ErrInvalidCounterpartyRole
	}

	return ps.saver.Update(ctx, id, name, role)	
}

func (ps *CounterpartyService) Delete(ctx context.Context, id int) (error) {
	return ps.saver.Delete(ctx, id)
}