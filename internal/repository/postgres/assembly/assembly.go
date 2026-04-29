package assembly

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

type AssemblyRepository struct {
	db QueryExecutor
}

func New(db QueryExecutor) *AssemblyRepository {
	return &AssemblyRepository{db: db}
}

func (r *AssemblyRepository) Create(ctx context.Context, vo entity.UpsertAssemblyVO) (int, error) {
	var id int
	err := r.db.QueryRow(ctx, `
		INSERT INTO assemblies (name, output_product_id, output_quantity, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id`,
		vo.Name,
		vo.OutputProductID,
		vo.OutputQuantity,
	).Scan(&id)

	if err != nil {
		log.Printf("AssemblyRepository::Create Error - %v", err)
		return 0, err
	}
	return id, nil
}

func (r *AssemblyRepository) Update(ctx context.Context, id int, vo entity.UpsertAssemblyVO) error {
	_, err := r.db.Exec(ctx, `
		UPDATE assemblies 
		SET name = $1, output_product_id = $2, output_quantity = $3
		WHERE id = $4`,
		vo.Name,
		vo.OutputProductID,
		vo.OutputQuantity,
		id,
	)
	if err != nil {
		log.Printf("AssemblyRepository::Update Error - %v", err)
		return err
	}
	return nil
}

func (r *AssemblyRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, "DELETE FROM assemblies WHERE id = $1", id)
	if err != nil {
		log.Printf("AssemblyRepository::Delete Error - %v", err)
		return err
	}
	return nil
}

func (r *AssemblyRepository) GetByID(ctx context.Context, id int) (*entity.Assembly, error) {
	var assembly entity.Assembly

	err := r.db.QueryRow(ctx, `
		SELECT id, name, output_product_id, output_quantity
		FROM assemblies WHERE id = $1`, id,
	).Scan(
		&assembly.ID,
		&assembly.Name,
		&assembly.OutputProductID,
		&assembly.OutputQuantity,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrAssemblyNotFound
		}
		log.Printf("AssemblyRepository::GetByID Error - %v", err)
		return nil, err
	}

	components, err := r.GetComponents(ctx, id)
	if err != nil {
		log.Printf("AssemblyRepository::GetByID GetComponents Error - %v", err)
		return nil, err
	}
	assembly.Components = components

	return &assembly, nil
}

func (r *AssemblyRepository) GetAll(ctx context.Context) ([]entity.Assembly, error) {
	var assemblies []entity.Assembly

	rows, err := r.db.Query(ctx, `
		SELECT id, name, output_product_id, output_quantity
		FROM assemblies ORDER BY id`)
	if err != nil {
		log.Printf("AssemblyRepository::GetAll Error - %v", err)
		return assemblies, err
	}
	defer rows.Close()

	for rows.Next() {
		var assembly entity.Assembly
		err := rows.Scan(
			&assembly.ID,
			&assembly.Name,
			&assembly.OutputProductID,
			&assembly.OutputQuantity,
		)
		if err != nil {
			log.Printf("AssemblyRepository::GetAll Scan Error - %v", err)
			continue
		}
		assemblies = append(assemblies, assembly)
	}

	for i := range assemblies {
		components, err := r.GetComponents(ctx, assemblies[i].ID)
		if err != nil {
			log.Printf("AssemblyRepository::GetAll GetComponents Error - %v", err)
			continue
		}
		assemblies[i].Components = components
	}

	return assemblies, rows.Err()
}

func (r *AssemblyRepository) AddComponent(ctx context.Context, assemblyID, productID int, quantity float64) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO assembly_components (assembly_id, product_id, quantity)
		VALUES ($1, $2, $3)`,
		assemblyID, productID, quantity,
	)
	if err != nil {
		log.Printf("AssemblyRepository::AddComponent Error - %v", err)
		return err
	}
	return nil
}

func (r *AssemblyRepository) GetComponents(ctx context.Context, assemblyID int) ([]entity.AssemblyComponent, error) {
	var components []entity.AssemblyComponent

	rows, err := r.db.Query(ctx, `
		SELECT id, assembly_id, product_id, quantity
		FROM assembly_components WHERE assembly_id = $1 ORDER BY id`, assemblyID,
	)
	if err != nil {
		log.Printf("AssemblyRepository::GetComponents Error - %v", err)
		return components, err
	}
	defer rows.Close()

	for rows.Next() {
		var comp entity.AssemblyComponent
		err := rows.Scan(
			&comp.ID,
			&comp.AssemblyID,
			&comp.ProductID,
			&comp.Quantity,
		)
		if err != nil {
			log.Printf("AssemblyRepository::GetComponents Scan Error - %v", err)
			continue
		}
		components = append(components, comp)
	}

	return components, rows.Err()
}

func (r *AssemblyRepository) DeleteComponents(ctx context.Context, assemblyID int) error {
	_, err := r.db.Exec(ctx, "DELETE FROM assembly_components WHERE assembly_id = $1", assemblyID)
	if err != nil {
		log.Printf("AssemblyRepository::DeleteComponents Error - %v", err)
		return err
	}
	return nil
}

func (r *AssemblyRepository) ReplaceComponents(ctx context.Context, assemblyID int, components []struct {
	ProductID int
	Quantity  float64
}) error {
	if err := r.DeleteComponents(ctx, assemblyID); err != nil {
		return err
	}

	for _, comp := range components {
		if err := r.AddComponent(ctx, assemblyID, comp.ProductID, comp.Quantity); err != nil {
			return err
		}
	}
	return nil
}