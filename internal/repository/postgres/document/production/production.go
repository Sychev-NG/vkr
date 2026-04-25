package production

import (
	"context"
	"errors"
	"log"
	"time"

	"vkr/internal/entity/document/production"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type QueryExecutor interface {
    QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
    Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
    Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type ProductionRepository struct {
    db QueryExecutor
}

func New(db QueryExecutor) *ProductionRepository {
    return &ProductionRepository{db: db}
}

func (r *ProductionRepository) Add(ctx context.Context, vo production.UpsertProductionDocumentVO) (*production.ProductionDocument, error) {
    documentId, err := r.createDocument(ctx, vo.WarehouseID, vo.RecipeID, vo.Quantity)
    if err != nil {
        log.Printf("ProductionRepository::Add createDocument Error - %v", err)
        return nil, err
    }

    result, err := r.GetById(ctx, documentId)
    if err != nil {
        log.Printf("ProductionRepository::Add GetById Error - %v", err)
        return nil, err
    }

    return result, nil
}

func (r *ProductionRepository) createDocument(ctx context.Context, warehouse_id, recipe_id int, quantity float32) (int, error) {
    var documentId int
    
    err := r.db.QueryRow(ctx, `
        INSERT INTO production_docs (warehouse_id, recipe_id, quantity, date) 
        VALUES ($1, $2, $3, $4) 
        RETURNING id`,
        warehouse_id,
        recipe_id,
        quantity,
        time.Now().UTC(),
    ).Scan(&documentId)

    if err != nil {
        log.Printf("ProductionRepository::createDocument Error - %v", err)
        return 0, err
    }

    return documentId, nil
}

func (r *ProductionRepository) GetById(ctx context.Context, id int) (*production.ProductionDocument, error) {
    var doc production.ProductionDocument

    err := r.db.QueryRow(ctx, `
        SELECT id, warehouse_id, recipe_id, quantity, date 
        FROM production_docs 
        WHERE id = $1`, id,
    ).Scan(
        &doc.ID,
        &doc.WarehouseID,
        &doc.RecipeID,
        &doc.Quantity,
        &doc.Date,
    )

    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, production.ErrProductionDocumentNotFound
        }
        log.Printf("ProductionRepository::GetById Error - %v", err)
        return nil, err
    }

    return &doc, nil
}

func (r *ProductionRepository) GetAll(ctx context.Context) ([]production.ProductionDocument, error) {
    var docs []production.ProductionDocument

    rows, err := r.db.Query(ctx, `
        SELECT id, warehouse_id, recipe_id, quantity, date 
        FROM production_docs
        ORDER BY date DESC`,
    )
    if err != nil {
        log.Printf("ProductionRepository::GetAll Error - %v", err)
        return docs, err
    }
    defer rows.Close()

    for rows.Next() {
        var doc production.ProductionDocument
        err := rows.Scan(
            &doc.ID,
            &doc.WarehouseID,
            &doc.RecipeID,
            &doc.Quantity,
            &doc.Date,
        )
        if err != nil {
            log.Printf("ProductionRepository::GetAll Scan Error - %v", err)
            return docs, err
        }

        docs = append(docs, doc)
    }

    return docs, rows.Err()
}