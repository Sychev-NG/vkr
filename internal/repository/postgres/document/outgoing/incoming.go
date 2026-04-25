package outgoing

import (
	"context"
	"errors"
	"log"
	"time"

	"vkr/internal/entity/document/outgoing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type QueryExecutor interface {
    QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
    Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
    Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type OutgoingRepository struct {
    db QueryExecutor
}

func New(db QueryExecutor) *OutgoingRepository {
    return &OutgoingRepository{db: db}
}

func (r *OutgoingRepository) Add(ctx context.Context, vo outgoing.UpsertOutgoingDocumentVO) (*outgoing.OutgoingDocument, error) {
    documentId, err := r.createDocument(ctx, vo.WarehouseID, vo.CounterPartyID)
    if err != nil {
        log.Printf("OutgoingRepository::Add createDocument Error - %v", err)
        return nil, err
    }

    for _, item := range vo.Items {
        _, err := r.createDocumentItem(ctx, documentId, item.FinishedMaterialID, item.Quantity, item.Price)
        if err != nil {
            log.Printf("OutgoingRepository::Add createDocumentItem Error - %v", err)
            return nil, err
        }
    }

    result, err := r.GetById(ctx, documentId)
    if err != nil {
        log.Printf("OutgoingRepository::Add GetById Error - %v", err)
        return nil, err
    }

    return result, nil
}

func (r *OutgoingRepository) createDocument(ctx context.Context, warehouse_id, cpunterparty_id int) (int, error) {
    var documentId int
    
    err := r.db.QueryRow(ctx, `
        INSERT INTO outgoing_docs (counterparty_id, warehouse_id, date) 
        VALUES ($1, $2, $3) 
        RETURNING id`,
        cpunterparty_id,
        warehouse_id,
        time.Now().UTC(),
    ).Scan(&documentId)

    if err != nil {
        log.Printf("OutgoingRepository::createDocument Error - %v", err)
        return 0, err
    }

    return documentId, nil
}

func (r *OutgoingRepository) createDocumentItem(ctx context.Context, documentId int, raw_material_id int, quantity, price float32) (int, error) {
    var itemId int
 
    err := r.db.QueryRow(ctx, `
        INSERT INTO outgoing_doc_items (document_id, product_id, quantity, price) 
        VALUES ($1, $2, $3, $4) 
        RETURNING id`,
        documentId,
        raw_material_id,
        quantity,
        price,
    ).Scan(&itemId)

    if err != nil {
        log.Printf("OutgoingRepository::createDocumentItem Error - %v", err)
        return 0, err
    }

    return itemId, nil
}

func (r *OutgoingRepository) GetById(ctx context.Context, id int) (*outgoing.OutgoingDocument, error) {
    var doc outgoing.OutgoingDocument

    err := r.db.QueryRow(ctx, `
        SELECT id, counterparty_id, warehouse_id, date 
        FROM outgoing_docs 
        WHERE id = $1`, id,
    ).Scan(
        &doc.ID,
        &doc.CounterPartyID,
        &doc.WarehouseID,
        &doc.Date,
    )

    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, outgoing.ErrOutgoingDocumentNotFound
        }
        log.Printf("OutgoingRepository::GetById Error - %v", err)
        return nil, err
    }

    items, err := r.getItems(ctx, doc.ID)
    if err != nil {
        log.Printf("OutgoingRepository::GetById getItems Error - %v", err)
        return nil, err
    }

    doc.Items = items

    return &doc, nil
}

func (r *OutgoingRepository) getItems(ctx context.Context, documentId int) ([]outgoing.OutgoingDocumentItem, error) {
    var items []outgoing.OutgoingDocumentItem

    rows, err := r.db.Query(ctx, `
        SELECT id, product_id, quantity, price
        FROM outgoing_doc_items 
        WHERE document_id = $1
        ORDER BY id`, documentId,
    )

    if err != nil {
        log.Printf("OutgoingRepository::getItems Error - %v", err)
        return items, err
    }
    defer rows.Close()

    for rows.Next() {
        var item outgoing.OutgoingDocumentItem
        err := rows.Scan(
            &item.ID,
            &item.FinishedMaterialID,
            &item.Quantity,
            &item.Price,
        )
        
        if err != nil {
            log.Printf("OutgoingRepository::getItems Scan Error - %v", err)
            return items, err
        }
        items = append(items, item)
    }

    return items, rows.Err()
}

func (r *OutgoingRepository) GetAll(ctx context.Context) ([]outgoing.OutgoingDocument, error) {
    var docs []outgoing.OutgoingDocument

    rows, err := r.db.Query(ctx, `
        SELECT id, counterparty_id, warehouse_id, date 
        FROM outgoing_docs
		ORDER BY date DESC`,
    )
    if err != nil {
        log.Printf("OutgoingRepository::GetAll Error - %v", err)
        return docs, err
    }
    defer rows.Close()

    for rows.Next() {
        var doc outgoing.OutgoingDocument
        err := rows.Scan(
            &doc.ID,
            &doc.CounterPartyID,
            &doc.WarehouseID,
            &doc.Date,
        )
        if err != nil {
            log.Printf("OutgoingRepository::GetAll Scan Error - %v", err)
            return docs, err
        }

        docItems, err := r.getItems(ctx, doc.ID)
        if err != nil {
            log.Printf("OutgoingRepository::GetAll getItems Error - %v", err)
            return docs, err
        }
        doc.Items = docItems

        docs = append(docs, doc)
    }

    return docs, rows.Err()
}