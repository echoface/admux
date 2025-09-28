package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/echoface/admux/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load configuration based on RUN_TYPE
	cfg, err := config.LoadConfig("trackingserver")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 设置Gin模式
	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 创建Gin路由
	r := gin.Default()

	// 根端点
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ADMUX Tracking Server is running!",
			"version": "v1.0.0-dev",
			"run_type": cfg.RunType,
		})
	})

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"run_type": cfg.RunType,
		})
	})

	// Prometheus指标端点
	if cfg.Monitoring.Prometheus.Enabled {
		r.GET(cfg.Monitoring.Prometheus.Endpoint, gin.WrapH(promhttp.Handler()))
	}

	// Tracking事件端点
	r.POST("/track/event", handleTrackEvent)
	r.GET("/track/impression", handleImpressionTracking)
	r.GET("/track/click", handleClickTracking)
	r.POST("/track/conversion", handleConversionTracking)

	// 启动服务器
	address := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("ADMUX Tracking Server starting on %s (run_type: %s)", address, cfg.RunType)
	log.Printf("Health check: http://localhost:%s/health", address)

	if err := r.Run(address); err != nil {
		log.Fatal("Failed to start Tracking server:", err)
	}
}

// Tracking事件处理函数
func handleTrackEvent(c *gin.Context) {
	// TODO: 实现事件跟踪逻辑
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "Event tracked successfully",
	})
}

func handleImpressionTracking(c *gin.Context) {
	// TODO: 实现曝光跟踪逻辑
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "Impression tracked successfully",
	})
}

func handleClickTracking(c *gin.Context) {
	// TODO: 实现点击跟踪逻辑
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "Click tracked successfully",
	})
}

func handleConversionTracking(c *gin.Context) {
	// TODO: 实现转化跟踪逻辑
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "Conversion tracked successfully",
	})
}