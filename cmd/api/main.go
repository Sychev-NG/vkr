package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vkr/internal/config"

	pHandler "vkr/internal/handlers/product"
	pService "vkr/internal/service/product"
	pRepo "vkr/internal/repository/postgres/product"

	cpHandler "vkr/internal/handlers/counterparty"
	cpService "vkr/internal/service/counterparty"
	cpRepo "vkr/internal/repository/postgres/counterparty"

	"vkr/internal/storage/postgres"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbPool *pgxpool.Pool

var productHandler *pHandler.ProductHandler
var counterpartyHandler *cpHandler.CounterpartyHandler

func main() {
	cfg := config.MustLoad()
	log.Println("Config is loaded")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error

	dbPool, err = postgres.NewPool(ctx, *cfg)
	if err != nil {
		log.Fatal("Failed to init DB: %v", err)
	}
	log.Println("Connection to DB is set")

	productRepository := pRepo.New(dbPool)
	productService := pService.New(productRepository, productRepository)
	productHandler = pHandler.New(productService)

	counterpartyRepository := cpRepo.New(dbPool)
	counterpartyService := cpService.New(counterpartyRepository, counterpartyRepository)
	counterpartyHandler = cpHandler.New(counterpartyService)

	err = dbPool.Ping(ctx)
	fmt.Printf("Pinging DB %v", err)

	r := setupRouter()
	
	srv := http.Server{
		Addr: 		":"+cfg.API.Port,
		Handler:	r, 
	}
	
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failder to start API server: %v", err)
		}
	}()

	log.Println("Preparing gracefull shutdown")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down the app gracefully")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	if dbPool != nil {
		dbPool.Close()
		log.Println("Database is closed")
	}
}

func setupRouter() *gin.Engine {
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			ctx := context.Background()
			err := dbPool.Ping(ctx)
			if err != nil {
				c.Status(501)
				return				
			}

			c.Status(200)
			return
		})

		// Группа товаров
		productGroup := api.Group("/product")
		{
			productGroup.GET("/product", productHandler.List)
			productGroup.GET("/product/:id", productHandler.Get)
			productGroup.POST("/product", productHandler.Create)
			productGroup.PATCH("/product/:id", productHandler.Update)
			productGroup.DELETE("/product/:id", productHandler.Delete)
		}

        // Группа контрагентов
        counterpartyGroup := api.Group("/counterparty")
        {
            counterpartyGroup.GET("", counterpartyHandler.List)
            counterpartyGroup.GET("/:id", counterpartyHandler.Get)
            counterpartyGroup.POST("", counterpartyHandler.Create)
            counterpartyGroup.PATCH("/:id", counterpartyHandler.Update)
            counterpartyGroup.DELETE("/:id", counterpartyHandler.Delete)
        }
	}

	return router
}