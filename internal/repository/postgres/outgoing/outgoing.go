package outgoing

import (
	"context"
	"errors"
	"log"
	"time"

	outgoingEntity "vkr/internal/entity/document/outgoing"

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

func (r *OutgoingRepository) Create(ctx context.Context, vo outgoingEntity.UpsertOutgoingDocumentVO) (*outgoingEntity.OutgoingDocument, error) {
	documentID, err := r.createDocument(ctx, vo.WarehouseID, vo.CounterPartyID)
	if err != nil {
		log.Printf("OutgoingRepository::Create createDocument Error - %v", err)
		return nil, err
	}

	for _, item := range vo.Items {
		_, err := r.createDocumentItem(ctx, documentID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			log.Printf("OutgoingRepository::Create createDocumentItem Error - %v", err)
			return nil, err
		}
	}

	document, err := r.GetByID(ctx, documentID) 
	if err != nil {
		log.Printf("OutgoingRepository::Create GetByID Error - %v", err)
		return nil, err
	}

	return document, nil
}

func (r *OutgoingRepository) createDocument(ctx context.Context, warehouseID, counterpartyID int) (int, error) {
	var documentID int

	err := r.db.QueryRow(ctx, `
		INSERT INTO outgoing_docs (counterparty_id, warehouse_id, created_at) 
		VALUES ($1, $2, $3) 
		RETURNING id`,
		counterpartyID,
		warehouseID,
		time.Now().UTC(),
	).Scan(&documentID)

	if err != nil {
		log.Printf("OutgoingRepository::createDocument Error - %v", err)
		return 0, err
	}

	return documentID, nil
}

func (r *OutgoingRepository) createDocumentItem(ctx context.Context, documentID, productID int, quantity, price float64) (int, error) {
	var itemID int

	err := r.db.QueryRow(ctx, `
		INSERT INTO outgoing_items (document_id, product_id, quantity, price) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id`,
		documentID,
		productID,
		quantity,
		price,
	).Scan(&itemID)

	if err != nil {
		log.Printf("OutgoingRepository::createDocumentItem Error - %v", err)
		return 0, err
	}

	return itemID, nil
}

func (r *OutgoingRepository) SetUnitCost(ctx context.Context, doc_id, product_id int, unit_cost float64) (error) {
	var itemID int

	err := r.db.QueryRow(ctx, `
		UPDATE outgoing_items 
		SET unit_cost = $1  
		WHERE document_id = $2 AND product_id = $3 RETURNING id`,
		unit_cost,
		doc_id,
		product_id,
	).Scan(&itemID)

	if err != nil {
		log.Printf("OutgoingRepository::SetUnitCost Error - %v", err)
		return err
	}

	return nil
}

func (r *OutgoingRepository) GetByID(ctx context.Context, id int) (*outgoingEntity.OutgoingDocument, error) {
	var doc outgoingEntity.OutgoingDocument

	err := r.db.QueryRow(ctx, `
		SELECT id, counterparty_id, warehouse_id, created_at 
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
			return nil, outgoingEntity.ErrOutgoingDocumentNotFound
		}
		log.Printf("OutgoingRepository::GetByID Error - %v", err)
		return nil, err
	}

	items, err := r.getItems(ctx, doc.ID)
	if err != nil {
		log.Printf("OutgoingRepository::GetByID getItems Error - %v", err)
		return nil, err
	}

	doc.Items = items

	return &doc, nil
}

func (r *OutgoingRepository) getItems(ctx context.Context, documentID int) ([]outgoingEntity.OutgoingDocumentItem, error) {
	var items []outgoingEntity.OutgoingDocumentItem

	rows, err := r.db.Query(ctx, `
		SELECT id, product_id, quantity, price, unit_cost
		FROM outgoing_items 
		WHERE document_id = $1
		ORDER BY id`, documentID,
	)

	if err != nil {
		log.Printf("OutgoingRepository::getItems Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item outgoingEntity.OutgoingDocumentItem
		var unitCost *float64
		err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.Quantity,
			&item.Price,
			&unitCost,
		)

		if unitCost != nil {
			item.UnitCost = *unitCost
		}

		if err != nil {
			log.Printf("OutgoingRepository::getItems Scan Error - %v", err)
			return items, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *OutgoingRepository) GetAll(ctx context.Context) ([]outgoingEntity.OutgoingDocument, error) {
	var docs []outgoingEntity.OutgoingDocument

	rows, err := r.db.Query(ctx, `
		SELECT id, counterparty_id, warehouse_id, created_at 
		FROM outgoing_docs
		ORDER BY created_at DESC`,
	)
	if err != nil {
		log.Printf("OutgoingRepository::GetAll Error - %v", err)
		return docs, err
	}
	defer rows.Close()

	for rows.Next() {
		var doc outgoingEntity.OutgoingDocument
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