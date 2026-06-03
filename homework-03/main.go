package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create a Gin router with default middleware (logger and recovery)
	r := gin.Default()

	// Define a simple GET endpoint
	r.GET("/health", func(c *gin.Context) {
		// Return JSON response
		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})
	})

	if err := r.Run("0.0.0.0:8000"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
