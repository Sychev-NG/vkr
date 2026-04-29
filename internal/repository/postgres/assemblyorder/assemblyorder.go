package assemblyorder

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	docAssembly "vkr/internal/entity/document/assembly"
)

type QueryExecutor interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type AssemblyOrderRepository struct {
	db QueryExecutor
}

func New(db QueryExecutor) *AssemblyOrderRepository {
	return &AssemblyOrderRepository{db: db}
}

func (r *AssemblyOrderRepository) Create(ctx context.Context, vo docAssembly.UpsertAssemblyOrderVO) (*docAssembly.AssemblyOrder, error) {
	var id int
	now := time.Now().UTC()
	
	err := r.db.QueryRow(ctx, `
		INSERT INTO assembly_orders (assembly_id, warehouse_id, quantity_to_build, doc_date, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		vo.AssemblyID,
		vo.WarehouseID,
		vo.QuantityToBuild,
		now,
		now,
	).Scan(&id)

	if err != nil {
		log.Printf("AssemblyOrderRepository::Create Error - %v", err)
		return nil, err
	}

	result, err := r.GetByID(ctx, id)
	if err != nil {
		log.Printf("AssemblyOrderRepository::Create Error - %v", err)
		return nil, err
	}
	return result, nil
}

func (r *AssemblyOrderRepository) AddConsumption(ctx context.Context, orderID, productID int, quantity, unitCost, totalCost float64) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO assembly_order_consumptions (order_id, product_id, quantity, unit_cost, total_cost)
		VALUES ($1, $2, $3, $4, $5)`,
		orderID, productID, quantity, unitCost, totalCost,
	)
	if err != nil {
		log.Printf("AssemblyOrderRepository::AddConsumption Error - %v", err)
		return err
	}
	return nil
}

func (r *AssemblyOrderRepository) AddOutput(ctx context.Context, orderID, productID int, quantity, unitCost, totalCost float64) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO assembly_order_outputs (order_id, product_id, quantity, unit_cost, total_cost)
		VALUES ($1, $2, $3, $4, $5)`,
		orderID, productID, quantity, unitCost, totalCost,
	)
	if err != nil {
		log.Printf("AssemblyOrderRepository::AddOutput Error - %v", err)
		return err
	}
	return nil
}

func (r *AssemblyOrderRepository) GetByID(ctx context.Context, id int) (*docAssembly.AssemblyOrder, error) {
	var order docAssembly.AssemblyOrder
	var docDate, createdAt time.Time

	err := r.db.QueryRow(ctx, `
		SELECT id, assembly_id, warehouse_id, quantity_to_build, doc_date, created_at
		FROM assembly_orders WHERE id = $1`, id,
	).Scan(
		&order.ID,
		&order.AssemblyID,
		&order.WarehouseID,
		&order.QuantityToBuild,
		&docDate,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, docAssembly.ErrAssemblyOrderNotFound
		}
		log.Printf("AssemblyOrderRepository::GetByID Error - %v", err)
		return nil, err
	}
	order.Date = docDate

	consumptions, err := r.GetConsumptionsByOrderID(ctx, id)
	if err != nil {
		log.Printf("AssemblyOrderRepository::GetByID GetConsumptions Error - %v", err)
		return nil, err
	}
	order.Consumptions = consumptions

	outputs, err := r.GetOutputsByOrderID(ctx, id)
	if err != nil {
		log.Printf("AssemblyOrderRepository::GetByID GetOutputs Error - %v", err)
		return nil, err
	}
	order.Outputs = outputs

	return &order, nil
}

func (r *AssemblyOrderRepository) GetConsumptionsByOrderID(ctx context.Context, orderID int) ([]docAssembly.AssemblyOrderConsumption, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, order_id, product_id, quantity, unit_cost, total_cost
		FROM assembly_order_consumptions WHERE order_id = $1 ORDER BY id`, orderID,
	)
	if err != nil {
		log.Printf("AssemblyOrderRepository::GetConsumptionsByOrderID Error - %v", err)
		return nil, err
	}
	defer rows.Close()

	var consumptions []docAssembly.AssemblyOrderConsumption
	for rows.Next() {
		var c docAssembly.AssemblyOrderConsumption
		err := rows.Scan(
			&c.ID,
			&c.OrderID,
			&c.ProductID,
			&c.Quantity,
			&c.UnitCost,
			&c.TotalCost,
		)
		if err != nil {
			log.Printf("AssemblyOrderRepository::GetConsumptionsByOrderID Scan Error - %v", err)
			continue
		}
		consumptions = append(consumptions, c)
	}

	return consumptions, rows.Err()
}

func (r *AssemblyOrderRepository) GetOutputsByOrderID(ctx context.Context, orderID int) ([]docAssembly.AssemblyOrderOutput, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, order_id, product_id, quantity, unit_cost, total_cost
		FROM assembly_order_outputs WHERE order_id = $1 ORDER BY id`, orderID,
	)
	if err != nil {
		log.Printf("AssemblyOrderRepository::GetOutputsByOrderID Error - %v", err)
		return nil, err
	}
	defer rows.Close()

	var outputs []docAssembly.AssemblyOrderOutput
	for rows.Next() {
		var o docAssembly.AssemblyOrderOutput
		err := rows.Scan(
			&o.ID,
			&o.OrderID,
			&o.ProductID,
			&o.Quantity,
			&o.UnitCost,
			&o.TotalCost,
		)
		if err != nil {
			log.Printf("AssemblyOrderRepository::GetOutputsByOrderID Scan Error - %v", err)
			continue
		}
		outputs = append(outputs, o)
	}

	return outputs, rows.Err()
}

func (r *AssemblyOrderRepository) GetAll(ctx context.Context) ([]docAssembly.AssemblyOrder, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, assembly_id, warehouse_id, quantity_to_build, doc_date, created_at
		FROM assembly_orders ORDER BY id DESC`)
	if err != nil {
		log.Printf("AssemblyOrderRepository::GetAll Error - %v", err)
		return nil, err
	}
	defer rows.Close()

	var orders []docAssembly.AssemblyOrder
	for rows.Next() {
		var order docAssembly.AssemblyOrder
		var docDate, createdAt time.Time
		
		err := rows.Scan(
			&order.ID,
			&order.AssemblyID,
			&order.WarehouseID,
			&order.QuantityToBuild,
			&docDate,
			&createdAt,
		)
		if err != nil {
			log.Printf("AssemblyOrderRepository::GetAll Scan Error - %v", err)
			continue
		}
		order.Date = docDate
		
		orders = append(orders, order)
	}

	return orders, rows.Err()
}