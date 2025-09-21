package main

import (
	"log"
	"net/http"

	"github.com/echoface/admux/internal/adxserver"
	"github.com/echoface/admux/internal/handler"
	"github.com/echoface/admux/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Initialize ADX server with pipeline
	adxServer := adxserver.NewAdxServer()
	bidHandler := handler.NewBidHandler(adxServer)

	r := gin.Default()

	// Add Prometheus metrics middleware
	r.Use(middleware.PrometheusMetrics())

	// Health check endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ADX Server is running!",
		})
	})

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// RTB bid endpoint
	r.POST("/bid/rtb/v1", bidHandler.HandleBidRequest)

	log.Println("ADX Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start ADX server:", err)
	}
}

