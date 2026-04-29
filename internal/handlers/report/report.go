package report

import (
	"context"
	"net/http"
	"time"
	"vkr/internal/entity/report"

	"github.com/gin-gonic/gin"
)

/*

[
	{
		"doсument_id": 102,
		"date": "2024-02-10",
		"counterparty_id": 5,
		"counterparty_name": "ООО 'Магнит'",
		"items": [
			{
				"product_id": 3,
				"product_name": "Хлеб пшеничный",
				"product_unit": "kg",
				"quantity": 150.00,
				"selling_price": 85.00,
				"unit_cost": 48.00,
				"revenue": 12750.00,
				"cogs": 7200.00,
				"profit": 5550.00
			}
		],
		"revenue": 12750.00,
		"cogs": 7200.00,
		"profit": 5550.00 // Валовая прибыльь (Общая прибыль - Себестоимость покупок)
	}
]

*/
type COGSDocument struct {
	DocumentID       int      		`json:"document_id"`
	Date             time.Time   		`json:"date"`
	CounterpartyID   int      		`json:"counterparty_id"`
	CounterpartyName string   		`json:"counterparty_name"`
	Revenue          float64  		`json:"revenue"` // Общая прибыль
	Cogs             float64  		`json:"cogs"`    // Себестоимость
	Profit           float64  		`json:"profit"` // Валовая прибыль
	Items            []COGSItem   	`json:"items"`
}

type COGSItem struct {
	ProductID    int     `json:"product_id"`
	ProductName  string  `json:"product_name"`
	ProductUnit  string  `json:"product_unit"`
	Quantity     float64 `json:"quantity"`
	SellingPrice float64 `json:"selling_price"`
	UnitCost     float64 `json:"unit_cost"`
	Revenue      float64 `json:"revenue"`
	Cogs         float64 `json:"cogs"`
	Profit       float64 `json:"profit"`
}

type COGSReportService interface {
	Count(ctx context.Context) ([]report.COGSDocument, error)
}

type ReportHandler struct {
	cogsService COGSReportService
}

func New(s COGSReportService) *ReportHandler {
	return &ReportHandler{
		cogsService: s,
	}
}

func (h *ReportHandler) COGS(c *gin.Context) {
	// fromStr := c.Query("from")
	// toStr := c.Query("to")

	// from, err := time.Parse("2006-01-02", fromStr)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date, use YYYY-MM-DD"})
	// 	return
	// }

	// to, err := time.Parse("2006-01-02", toStr)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date, use YYYY-MM-DD"})
	// 	return
	// }
	// to = to.Add(24*time.Hour - time.Second)

	data, err := h.cogsService.Count(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get COGS report: " + err.Error()})
		return
	}

	if data == nil {
		c.JSON(http.StatusOK, []COGSDocument{})
		return
	}

	response := make([]COGSDocument, len(data))

	for i, doc := range data {
		items := make([]COGSItem, len(doc.Items))
		for j, item := range doc.Items {
			items[j] = COGSItem{
				ProductID:    item.ProductID,
				ProductName:  item.ProductName,
				ProductUnit:  item.ProductUnit,
				Quantity:     item.Quantity,
				SellingPrice: item.SellingPrice,
				UnitCost:     item.UnitCost,
				Revenue:      item.Revenue,
				Cogs:         item.Cogs,
				Profit:       item.Profit,
			}
		}
		
		response[i] = COGSDocument{
			DocumentID:       doc.DocumentID,
			Date:             doc.Date,
			CounterpartyID:   doc.CounterpartyID,
			CounterpartyName: doc.CounterpartyName,
			Revenue:          doc.Revenue,
			Cogs:             doc.Cogs,
			Profit:           doc.Profit,
			Items:            items,
		}
	}

	c.JSON(http.StatusOK, response)
}