package counterparty

import (
	"context"
	"errors"
	"log"
	"vkr/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CounterpartyRepository struct {
	pool *pgxpool.Pool
}

func New(db *pgxpool.Pool) *CounterpartyRepository {
	return &CounterpartyRepository{pool: db}
}

func (pr *CounterpartyRepository) Add(ctx context.Context, name, role string) (*entity.Counterparty, error) {
	var item entity.Counterparty

	err := pr.pool.QueryRow(
		ctx, 
		"INSERT INTO counterparties (name, role) VALUES ($1, $2) RETURNING id, name, role", 
		name, 
		role, 
	).Scan(
		&item.ID,
		&item.Name,
		&item.Role,
	)

	if err != nil {
		log.Printf("CounterpartyRepository::Add Error - %v", err)
		return nil, err
	}

	return &item, nil
}

func (pr *CounterpartyRepository) Update(ctx context.Context, id int, name, role string) (*entity.Counterparty, error) {
	var item entity.Counterparty

	err := pr.pool.QueryRow(
		ctx, 
		"UPDATE counterparties SET name=$1, role=$2 WHERE id=$3 RETURNING id, name, role", 
		name, 
		role,
		id,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Role,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrCounterpartyNotFound
		}

		log.Printf("CounterpartyRepository::Update Error - %v", err)
		return nil, err
	}

	return &item, nil
}

func (pr *CounterpartyRepository) Delete(ctx context.Context, id int) error {
	_, err :=pr.pool.Exec(ctx, "DELETE FROM counterparties WHERE id=$1", id)
	if err != nil {
		log.Printf("CounterpartyRepository::Delete Error - %v", err)
	}
	return err
}

func (pr *CounterpartyRepository) GetById(ctx context.Context, id int) (*entity.Counterparty, error) {
    var item entity.Counterparty

    err := pr.pool.QueryRow(ctx, "SELECT id, name, role FROM counterparties WHERE id = $1", id).Scan(
		&item.ID, 
		&item.Name, 
		&item.Role,
	)
    
    if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrCounterpartyNotFound
		}

		log.Printf("CounterpartyRepository::GetById Error - %v", err)

        return nil, err
    }
        
    return &item, nil
}

func (pr *CounterpartyRepository) GetAll(ctx context.Context) ([]entity.Counterparty, error) {
	var items []entity.Counterparty

	rows, err := pr.pool.Query(ctx, "SELECT id, name, role FROM counterparties")
	if err != nil {
		log.Printf("CounterpartyRepository::GetAll Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Counterparty
		rows.Scan(
			&item.ID, 
			&item.Name, 
			&item.Role, 
		)
		items = append(items, item)
	}

	return items, err
}
