package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"vkr/internal/config"
)

func main() {
	cfg := config.MustLoad()

	log.Println("API_PORT: " + cfg.API.Port)

	r := setupRouter()	
	if err := r.Run(":" + cfg.API.Port); err != nil {
		log.Fatal(err)
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