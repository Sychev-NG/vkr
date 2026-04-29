package incoming

import (
	"context"
	"errors"
	"log"
	"time"

	incomingEntity "vkr/internal/entity/document/incoming"

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

func (r *IncomingRepository) Create(ctx context.Context, vo incomingEntity.UpsertIncomingDocumentVO) (*incomingEntity.IncomingDocument, error) {
	documentID, err := r.createDocument(ctx, vo.WarehouseID, vo.CounterPartyID)
	if err != nil {
		log.Printf("IncomingRepository::Create createDocument Error - %v", err)
		return nil, err
	}

	for _, item := range vo.Items {
		_, err := r.createDocumentItem(ctx, documentID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			log.Printf("IncomingRepository::Create createDocumentItem Error - %v", err)
			return nil, err
		}
	}

	document, err := r.GetByID(ctx, documentID) 
	if err != nil {
		log.Printf("IncomingRepository::Create GetByID Error - %v", err)
		return nil, err
	}

	return document, nil
}

func (r *IncomingRepository) createDocument(ctx context.Context, warehouseID, counterpartyID int) (int, error) {
	var documentID int

	err := r.db.QueryRow(ctx, `
		INSERT INTO incoming_docs (counterparty_id, warehouse_id, created_at) 
		VALUES ($1, $2, $3) 
		RETURNING id`,
		counterpartyID,
		warehouseID,
		time.Now().UTC(),
	).Scan(&documentID)

	if err != nil {
		log.Printf("IncomingRepository::createDocument Error - %v", err)
		return 0, err
	}

	return documentID, nil
}

func (r *IncomingRepository) createDocumentItem(ctx context.Context, documentID, productID int, quantity, price float64) (int, error) {
	var itemID int

	err := r.db.QueryRow(ctx, `
		INSERT INTO incoming_items (document_id, product_id, quantity, price) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id`,
		documentID,
		productID,
		quantity,
		price,
	).Scan(&itemID)

	if err != nil {
		log.Printf("IncomingRepository::createDocumentItem Error - %v", err)
		return 0, err
	}

	return itemID, nil
}

func (r *IncomingRepository) GetByID(ctx context.Context, id int) (*incomingEntity.IncomingDocument, error) {
	var doc incomingEntity.IncomingDocument

	err := r.db.QueryRow(ctx, `
		SELECT id, counterparty_id, warehouse_id, created_at 
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
			return nil, incomingEntity.ErrIncomingDocumentNotFound
		}
		log.Printf("IncomingRepository::GetByID Error - %v", err)
		return nil, err
	}

	items, err := r.getItems(ctx, doc.ID)
	if err != nil {
		log.Printf("IncomingRepository::GetByID getItems Error - %v", err)
		return nil, err
	}

	doc.Items = items

	return &doc, nil
}

func (r *IncomingRepository) getItems(ctx context.Context, documentID int) ([]incomingEntity.IncomingDocumentItem, error) {
	var items []incomingEntity.IncomingDocumentItem

	rows, err := r.db.Query(ctx, `
		SELECT id, product_id, quantity, price
		FROM incoming_items 
		WHERE document_id = $1
		ORDER BY id`, documentID,
	)

	if err != nil {
		log.Printf("IncomingRepository::getItems Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item incomingEntity.IncomingDocumentItem
		err := rows.Scan(
			&item.ID,
			&item.ProductID,
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

func (r *IncomingRepository) GetAll(ctx context.Context) ([]incomingEntity.IncomingDocument, error) {
	var docs []incomingEntity.IncomingDocument

	rows, err := r.db.Query(ctx, `
		SELECT id, counterparty_id, warehouse_id, created_at 
		FROM incoming_docs
		ORDER BY created_at DESC`,
	)
	if err != nil {
		log.Printf("IncomingRepository::GetAll Error - %v", err)
		return docs, err
	}
	defer rows.Close()

	for rows.Next() {
		var doc incomingEntity.IncomingDocument
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