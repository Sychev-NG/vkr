package counterparty

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

type CounterpartyRepository struct {
	db QueryExecutor
}

func New(db QueryExecutor) *CounterpartyRepository {
	return &CounterpartyRepository{db: db}
}

func (r *CounterpartyRepository) Create(ctx context.Context, name, role string) (*entity.Counterparty, error) {
	var item entity.Counterparty

	err := r.db.QueryRow(
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
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, entity.ErrCounterpartyDuplicateFound
		}
		log.Printf("CounterpartyRepository::Create Error - %v", err)
		return nil, err
	}

	return &item, nil
}

func (r *CounterpartyRepository) Update(ctx context.Context, id int, name, role string) (*entity.Counterparty, error) {
	var item entity.Counterparty

	err := r.db.QueryRow(
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
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, entity.ErrCounterpartyDuplicateFound
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrCounterpartyNotFound
		}
		log.Printf("CounterpartyRepository::Update Error - %v", err)
		return nil, err
	}

	return &item, nil
}

func (r *CounterpartyRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, "DELETE FROM counterparties WHERE id=$1", id)
	if err != nil {
		log.Printf("CounterpartyRepository::Delete Error - %v", err)
	}
	return err
}

func (r *CounterpartyRepository) GetByID(ctx context.Context, id int) (*entity.Counterparty, error) {
	var item entity.Counterparty

	err := r.db.QueryRow(ctx, "SELECT id, name, role FROM counterparties WHERE id = $1", id).Scan(
		&item.ID,
		&item.Name,
		&item.Role,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrCounterpartyNotFound
		}
		log.Printf("CounterpartyRepository::GetByID Error - %v", err)
		return nil, err
	}

	return &item, nil
}

func (r *CounterpartyRepository) GetAll(ctx context.Context) ([]entity.Counterparty, error) {
	var items []entity.Counterparty

	rows, err := r.db.Query(ctx, "SELECT id, name, role FROM counterparties ORDER BY id")
	if err != nil {
		log.Printf("CounterpartyRepository::GetAll Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Counterparty
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Role,
		); err != nil {
			log.Printf("CounterpartyRepository::GetAll Scan Error - %v", err)
			continue
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *CounterpartyRepository) GetByRole(ctx context.Context, role string) ([]entity.Counterparty, error) {
	var items []entity.Counterparty

	rows, err := r.db.Query(ctx, "SELECT id, name, role FROM counterparties WHERE role = $1 ORDER BY id", role)
	if err != nil {
		log.Printf("CounterpartyRepository::GetByRole Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Counterparty
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Role,
		); err != nil {
			log.Printf("CounterpartyRepository::GetByRole Scan Error - %v", err)
			continue
		}
		items = append(items, item)
	}

	return items, nil
}