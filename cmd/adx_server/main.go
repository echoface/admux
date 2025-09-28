package main

import (
	"log"
	"net/http"

	"github.com/echoface/admux/internal/adxserver"
	"github.com/echoface/admux/internal/config"
	"github.com/echoface/admux/internal/handler"
	"github.com/echoface/admux/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Initialize application context
	cfg := config.NewDefaultConfig()
	appCtx := adxserver.NewAppContext(cfg)

	// Initialize ADX server with pipeline
	adxServer := adxserver.NewAdxServer(appCtx)
	bidHandler := handler.NewBidHandler(adxServer, appCtx)

	// Setup routes using the router from app context
	r := appCtx.Router

	// Add Prometheus metrics middleware
	r.Use(middleware.PrometheusMetrics())

	// Health check endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ADX Server is running!",
			"healthy": appCtx.IsApplicationHealthy(),
		})
	})

	// Health status endpoint
	r.GET("/health", func(c *gin.Context) {
		if appCtx.IsApplicationHealthy() {
			c.JSON(http.StatusOK, gin.H{
				"status": "healthy",
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
			})
		}
	})

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// RTB bid endpoint
	r.POST("/bid/rtb/v1", bidHandler.HandleBidRequest)

	log.Println("ADX Server starting on :8080")

	// Use the HTTP server from app context
	if err := appCtx.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start ADX server:", err)
	}
}

