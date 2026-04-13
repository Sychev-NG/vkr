package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vkr/internal/config"
	"vkr/internal/storage/postgres"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.MustLoad()
	log.Println("Config is loaded")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbPool, err := postgres.NewPool(ctx, *cfg)
	if err != nil {
		log.Fatal("Failed to init DB: %v", err)
	}
	log.Println("Connection to DB is set")

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
			c.Status(200)
			return
		})
	}

	return router
}