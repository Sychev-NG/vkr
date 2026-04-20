package warehouse

import (
	"context"
	"errors"
	"log"
	"vkr/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WarehouseRepository struct {
	pool *pgxpool.Pool
}

func New(db *pgxpool.Pool) *WarehouseRepository {
	return &WarehouseRepository{pool: db}
}

func (pr *WarehouseRepository) Add(ctx context.Context, name, address string) (*entity.Warehouse, error) {
	var item entity.Warehouse

	err := pr.pool.QueryRow(
		ctx, 
		"INSERT INTO warehouses (name, address) VALUES ($1, $2) RETURNING id, name, address", 
		name, 
		address, 
	).Scan(
		&item.ID,
		&item.Name,
		&item.Address,
	)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" { // Psql 23505 - duplicate error code
			return nil, entity.ErrWarehouseDuplicateFound
		}

		log.Printf("WarehouseRepository::Add Error - %v", err)
		return nil, err
	}

	return &item, nil
}

func (pr *WarehouseRepository) Update(ctx context.Context, id int, name, address string) (*entity.Warehouse, error) {
	var item entity.Warehouse

	err := pr.pool.QueryRow(
		ctx, 
		"UPDATE warehouses SET name=$1, address=$2 WHERE id=$3 RETURNING id, name, address", 
		name, 
		address,
		id,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Address,
	)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" { // Psql 23505 - duplicate error code
			return nil, entity.ErrWarehouseDuplicateFound
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrWarehouseNotFound
		}

		log.Printf("WarehouseRepository::Update Error - %v", err)
		return nil, err
	}

	return &item, nil
}

func (pr *WarehouseRepository) Delete(ctx context.Context, id int) error {
	_, err :=pr.pool.Exec(ctx, "DELETE FROM warehouses WHERE id=$1", id)
	if err != nil {
		log.Printf("WarehouseRepository::Delete Error - %v", err)
	}
	return err
}

func (pr *WarehouseRepository) GetById(ctx context.Context, id int) (*entity.Warehouse, error) {
    var item entity.Warehouse

    err := pr.pool.QueryRow(ctx, "SELECT id, name, address FROM warehouses WHERE id = $1", id).Scan(
		&item.ID, 
		&item.Name, 
		&item.Address,
	)
    
    if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrWarehouseNotFound
		}

		log.Printf("WarehouseRepository::GetById Error - %v", err)

        return nil, err
    }
        
    return &item, nil
}

func (pr *WarehouseRepository) GetAll(ctx context.Context) ([]entity.Warehouse, error) {
	var items []entity.Warehouse

	rows, err := pr.pool.Query(ctx, "SELECT id, name, address FROM warehouses")
	if err != nil {
		log.Printf("WarehouseRepository::GetAll Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Warehouse
		rows.Scan(
			&item.ID, 
			&item.Name, 
			&item.Address, 
		)
		items = append(items, item)
	}

	return items, err
}
