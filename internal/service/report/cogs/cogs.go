package cogs

import (
	"context"
	"log"

	"vkr/internal/entity"
	"vkr/internal/entity/report"
	oRepo "vkr/internal/repository/postgres/outgoing"
)

type TxManager interface {
	RunInTx(ctx context.Context, fn func(txCtx context.Context) error) error
}

type ProductProvider interface {
	GetByID(ctx context.Context, id int) (*entity.Product, error)
}

type CounterpartyProvider interface {
	GetByID(ctx context.Context, id int) (*entity.Counterparty, error)
}

type RepoFactory interface {
	NewOutgoingRepository(ctx context.Context) *oRepo.OutgoingRepository
}

type COGSReportService struct {
	txManager 				TxManager
	f						RepoFactory
	productProvider			ProductProvider
	counterpartyProvider	CounterpartyProvider
}

func New(
	txManager TxManager, 
	f RepoFactory,
	pp ProductProvider,  
	cp CounterpartyProvider,
) *COGSReportService {
	return &COGSReportService{txManager, f, pp, cp}
}

func (s *COGSReportService) Count(ctx context.Context) ([]report.COGSDocument, error) {
	var result []report.COGSDocument

	outgoingRepo := s.f.NewOutgoingRepository(ctx)	
	outgoingDocumentCollection, err := outgoingRepo.GetAll(ctx)
	if err != nil {
		log.Printf("COGSReportService::Count outgoingRepo.GetAll Error - %v", err)
		return result, err
	}

	cogsCollection := make([]report.COGSDocument, len(outgoingDocumentCollection))

	for i, outgoingDocument := range outgoingDocumentCollection {
		cogsItemCollection := make([]report.COGSItem, len(outgoingDocument.Items))
		
		var totalRevenue float64
		var totalCogs float64
		var totalProfit float64
		
		for j, outgoingDocumentItem := range outgoingDocument.Items {

			product, err := s.productProvider.GetByID(ctx, outgoingDocumentItem.ProductID)
			if err != nil {
				return result, err
			}

			revenue := outgoingDocumentItem.Quantity * outgoingDocumentItem.Price
			cogs := outgoingDocumentItem.Quantity * outgoingDocumentItem.UnitCost
			profit := revenue - cogs
			
			totalRevenue += revenue
			totalCogs += cogs
			totalProfit += profit

			cogsItemCollection[j] = report.COGSItem{
				ProductID:     outgoingDocumentItem.ProductID,
				ProductName:   product.Name,
				ProductUnit:   product.Unit,
				Quantity:      outgoingDocumentItem.Quantity,
				SellingPrice:  outgoingDocumentItem.Price,
				UnitCost:      outgoingDocumentItem.UnitCost,
				Revenue:       revenue,
				Cogs:          cogs,
				Profit:        profit,
			}
		}	

		buyer, err := s.counterpartyProvider.GetByID(ctx, outgoingDocument.CounterPartyID)
		if err != nil {
			return result, err
		}

		cogsCollection[i] = report.COGSDocument{
			DocumentID:       outgoingDocument.ID,
			Date:             outgoingDocument.Date,
			CounterpartyID:   outgoingDocument.CounterPartyID,
			CounterpartyName: buyer.Name,
			Revenue:          totalRevenue,
			Cogs:             totalCogs,
			Profit:           totalProfit,
			Items:            cogsItemCollection,
		}
	}

	return cogsCollection, nil
}