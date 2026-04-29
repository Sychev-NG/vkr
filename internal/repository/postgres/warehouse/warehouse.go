package warehouse

import (
	"context"
	"errors"
	"log"
	"vkr/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type QueryExecutor interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type WarehouseRepository struct {
	db QueryExecutor
}

func New(db QueryExecutor) *WarehouseRepository {
	return &WarehouseRepository{db: db}
}

func (r *WarehouseRepository) Create(ctx context.Context, name, address string) (*entity.Warehouse, error) {
	var item entity.Warehouse

	err := r.db.QueryRow(
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
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, entity.ErrWarehouseDuplicateFound
		}
		log.Printf("WarehouseRepository::Create Error - %v", err)
		return nil, err
	}

	return &item, nil
}

func (r *WarehouseRepository) Update(ctx context.Context, id int, name, address string) (*entity.Warehouse, error) {
	var item entity.Warehouse

	err := r.db.QueryRow(
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
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
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

func (r *WarehouseRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, "DELETE FROM warehouses WHERE id=$1", id)
	if err != nil {
		log.Printf("WarehouseRepository::Delete Error - %v", err)
	}
	return err
}

func (r *WarehouseRepository) GetByID(ctx context.Context, id int) (*entity.Warehouse, error) {
	var item entity.Warehouse

	err := r.db.QueryRow(ctx, "SELECT id, name, address FROM warehouses WHERE id = $1", id).Scan(
		&item.ID,
		&item.Name,
		&item.Address,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrWarehouseNotFound
		}
		log.Printf("WarehouseRepository::GetByID Error - %v", err)
		return nil, err
	}

	return &item, nil
}

func (r *WarehouseRepository) GetAll(ctx context.Context) ([]entity.Warehouse, error) {
	var items []entity.Warehouse

	rows, err := r.db.Query(ctx, "SELECT id, name, address FROM warehouses ORDER BY id")
	if err != nil {
		log.Printf("WarehouseRepository::GetAll Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Warehouse
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Address,
		); err != nil {
			log.Printf("WarehouseRepository::GetAll Scan Error - %v", err)
			continue
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *WarehouseRepository) GetByIDs(ctx context.Context, ids []int) ([]entity.Warehouse, error) {
	if len(ids) == 0 {
		return []entity.Warehouse{}, nil
	}

	var items []entity.Warehouse
	rows, err := r.db.Query(ctx, "SELECT id, name, address FROM warehouses WHERE id = ANY($1)", ids)
	if err != nil {
		log.Printf("WarehouseRepository::GetByIDs Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Warehouse
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Address,
		); err != nil {
			log.Printf("WarehouseRepository::GetByIDs Scan Error - %v", err)
			continue
		}
		items = append(items, item)
	}

	return items, nil
}