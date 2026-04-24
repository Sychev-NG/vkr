package incoming

import (
	"context"
	"errors"
	"log"
	"time"

	"vkr/internal/entity/document/incoming"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type QueryExecutor interface {
    QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
    Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
    Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type IncomingRepository struct {
    db QueryExecutor
}

func New(db QueryExecutor) *IncomingRepository {
    return &IncomingRepository{db: db}
}

func (r *IncomingRepository) Add(ctx context.Context, vo incoming.UpsertIncomingDocumentVO) (*incoming.IncomingDocument, error) {
    documentId, err := r.createDocument(ctx, vo.WarehouseID, vo.CounterPartyID)
    if err != nil {
        log.Printf("IncomingRepository::Add createDocument Error - %v", err)
        return nil, err
    }

    for _, item := range vo.Items {
        _, err := r.createDocumentItem(ctx, documentId, item.RawMaterialID, item.Quantity, item.Price)
        if err != nil {
            log.Printf("IncomingRepository::Add createDocumentItem Error - %v", err)
            return nil, err
        }
    }

    result, err := r.GetById(ctx, documentId)
    if err != nil {
        log.Printf("IncomingRepository::Add GetById Error - %v", err)
        return nil, err
    }

    return result, nil
}

func (r *IncomingRepository) createDocument(ctx context.Context, warehouse_id, cpunterparty_id int) (int, error) {
    var documentId int
    
    err := r.db.QueryRow(ctx, `
        INSERT INTO incoming_docs (counterparty_id, warehouse_id, date) 
        VALUES ($1, $2, $3) 
        RETURNING id`,
        cpunterparty_id,
        warehouse_id,
        time.Now().UTC(),
    ).Scan(&documentId)

    if err != nil {
        log.Printf("IncomingRepository::createDocument Error - %v", err)
        return 0, err
    }

    return documentId, nil
}

func (r *IncomingRepository) createDocumentItem(ctx context.Context, documentId int, raw_material_id int, quantity, price float32) (int, error) {
    var itemId int

    err := r.db.QueryRow(ctx, `
        INSERT INTO incoming_doc_items (document_id, product_id, quantity, price) 
        VALUES ($1, $2, $3, $4) 
        RETURNING id`,
        documentId,
        raw_material_id,
        quantity,
        price,
    ).Scan(&itemId)

    if err != nil {
        log.Printf("IncomingRepository::createDocumentItem Error - %v", err)
        return 0, err
    }

    return itemId, nil
}

func (r *IncomingRepository) GetById(ctx context.Context, id int) (*incoming.IncomingDocument, error) {
    var doc incoming.IncomingDocument

    err := r.db.QueryRow(ctx, `
        SELECT id, counterparty_id, warehouse_id, date 
        FROM incoming_docs 
        WHERE id = $1`, id,
    ).Scan(
        &doc.ID,
        &doc.CounterPartyID,
        &doc.WarehouseID,
        &doc.Date,
    )

    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, incoming.ErrIncomingDocumentNotFound
        }
        log.Printf("IncomingRepository::GetById Error - %v", err)
        return nil, err
    }

    items, err := r.getItems(ctx, doc.ID)
    if err != nil {
        log.Printf("IncomingRepository::GetById getItems Error - %v", err)
        return nil, err
    }

    doc.Items = items

    return &doc, nil
}

func (r *IncomingRepository) getItems(ctx context.Context, documentId int) ([]incoming.IncomingDocumentItem, error) {
    var items []incoming.IncomingDocumentItem

    rows, err := r.db.Query(ctx, `
        SELECT id, product_id, quantity, price
        FROM incoming_doc_items 
        WHERE document_id = $1
        ORDER BY id`, documentId,
    )

    if err != nil {
        log.Printf("IncomingRepository::getItems Error - %v", err)
        return items, err
    }
    defer rows.Close()

    for rows.Next() {
        var item incoming.IncomingDocumentItem
        err := rows.Scan(
            &item.ID,
            &item.RawMaterialID,
            &item.Quantity,
            &item.Price,
        )
        
        if err != nil {
            log.Printf("IncomingRepository::getItems Scan Error - %v", err)
            return items, err
        }
        items = append(items, item)
    }

    return items, rows.Err()
}

func (r *IncomingRepository) GetAll(ctx context.Context) ([]incoming.IncomingDocument, error) {
    var docs []incoming.IncomingDocument

    rows, err := r.db.Query(ctx, `
        SELECT id, counterparty_id, warehouse_id, date 
        FROM incoming_docs
		ORDER BY date DESC`,
    )
    if err != nil {
        log.Printf("IncomingRepository::GetAll Error - %v", err)
        return docs, err
    }
    defer rows.Close()

    for rows.Next() {
        var doc incoming.IncomingDocument
        err := rows.Scan(
            &doc.ID,
            &doc.CounterPartyID,
            &doc.WarehouseID,
            &doc.Date,
        )
        if err != nil {
            log.Printf("IncomingRepository::GetAll Scan Error - %v", err)
            return docs, err
        }

        docItems, err := r.getItems(ctx, doc.ID)
        if err != nil {
            log.Printf("IncomingRepository::GetAll getItems Error - %v", err)
            return docs, err
        }
        doc.Items = docItems

        docs = append(docs, doc)
    }

    return docs, rows.Err()
}