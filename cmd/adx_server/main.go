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
	// Load configuration based on RUN_TYPE
	cfg, err := config.LoadConfig("adxserver")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize application context
	appCtx := adxserver.NewAppContext(cfg)

	// Initialize ADX server with pipeline
	adxServer := adxserver.NewAdxServer(appCtx)
	bidHandler := handler.NewBidHandler(adxServer, appCtx)

	// Initialize health handler
	healthHandler := handler.NewHealthHandler(appCtx, appCtx.GetMetricsRegistry())

	// Setup routes using the router from app context
	r := appCtx.Router

	// Add Prometheus metrics middleware
	r.Use(middleware.PrometheusMetrics())

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ADMUX ADX Server is running!",
			"version": "v1.0.0-dev",
			"healthy": appCtx.IsApplicationHealthy(),
		})
	})

	// Health check endpoints
	r.GET("/health", healthHandler.HealthCheck)
	r.GET("/health/live", healthHandler.LivenessProbe)
	r.GET("/health/ready", healthHandler.ReadinessProbe)

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// RTB bid endpoints
	r.POST("/bid/rtb/v1", bidHandler.HandleBidRequest)

	// SSP-specific endpoints
	r.POST("/bid/xiaomi", bidHandler.HandleBidRequest)
	r.POST("/bid/kuaishou", bidHandler.HandleBidRequest)
	r.POST("/bid/bytedance", bidHandler.HandleBidRequest)
	r.POST("/bid/maoyan", bidHandler.HandleBidRequest)

	log.Println("ADMUX ADX Server starting on port 8080")
	log.Printf("Health check: http://localhost:8080/health")
	log.Printf("Metrics: http://localhost:8080/metrics")

	// Use the HTTP server from app context
	if err := appCtx.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start ADX server:", err)
	}
}

