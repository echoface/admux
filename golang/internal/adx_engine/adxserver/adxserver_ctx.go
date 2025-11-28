package adxserver

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/echoface/admux/internal/adx_engine/config"
	"github.com/echoface/admux/internal/adx_engine/sspadapter"
	"github.com/echoface/admux/pkg/jsonx"
	"github.com/echoface/admux/pkg/logger"
)

// AdxServerContext represents the global application context for the ADX server
type AdxServerContext struct {
	// Configuration
	Config *config.ServerConfig

	// HTTP server
	HTTPServer *http.Server
	Router     *gin.Engine

	// Cache and storage
	// RedisClient *redis.Client // TODO: Add Redis client when implemented

	// Metrics
	MetricsRegistry *prometheus.Registry

	// SSP adapters (legacy)
	SSPAdapters map[string]SSPAdapter

	// SSP factory for new adapter architecture
	SSPFactory *sspadapter.SSPAdapterFactory

	// HTTP client for external API calls
	HTTPClient *http.Client

	// Context for graceful shutdown
	ShutdownCtx    context.Context
	ShutdownCancel context.CancelFunc

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Logger
	Logger logger.Logger

	// Health status
	IsHealthy bool
}

// SSPAdapter interface for SSP integration
type SSPAdapter interface {
	// Validate validates the SSP request
	Validate(ctx context.Context) (context.Context, error)
	// ConvertToOpenRTB converts SSP-specific request to OpenRTB
	ConvertToOpenRTB(data any) (any, error)
	// ConvertFromOpenRTB converts OpenRTB response to SSP-specific format
	ConvertFromOpenRTB(data any) (any, error)
	// GetSSPID returns the SSP identifier
	GetSSPID() string
}

// NewAppContext creates and initializes a new application context
func NewAppContext(cfg *config.ServerConfig) *AdxServerContext {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}
	fmt.Println("use config:", jsonx.Pretty(cfg))
	// Create context for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())

	// Initialize router
	router := gin.Default()

	// Create HTTP client with reasonable timeouts
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	appCtx := &AdxServerContext{
		Config:          cfg,
		HTTPServer:      httpServer,
		Router:          router,
		MetricsRegistry: prometheus.NewRegistry(),
		SSPAdapters:     make(map[string]SSPAdapter),
		HTTPClient:      httpClient,
		ShutdownCtx:     shutdownCtx,
		ShutdownCancel:  shutdownCancel,
		Logger:          logger.Default,
		IsHealthy:       true,
	}

	return appCtx
}

// RegisterSSPAdapter registers a new SSP adapter
func (ac *AdxServerContext) RegisterSSPAdapter(adapter SSPAdapter) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.SSPAdapters[adapter.GetSSPID()] = adapter
}

// GetSSPAdapter retrieves an SSP adapter by ID
func (ac *AdxServerContext) GetSSPAdapter(sspID string) (SSPAdapter, bool) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	adapter, exists := ac.SSPAdapters[sspID]
	return adapter, exists
}

// SetHealthStatus sets the overall health status of the application
func (ac *AdxServerContext) SetHealthStatus(healthy bool) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.IsHealthy = healthy
}

// IsApplicationHealthy returns the overall health status
func (ac *AdxServerContext) IsApplicationHealthy() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.IsHealthy
}

// Shutdown initiates graceful shutdown
func (ac *AdxServerContext) Shutdown() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.Logger.Info("Initiating graceful shutdown...")
	ac.IsHealthy = false

	if ac.ShutdownCancel != nil {
		ac.ShutdownCancel()
	}

	// Close HTTP client connections
	if ac.HTTPClient != nil {
		ac.HTTPClient.CloseIdleConnections()
	}

	ac.Logger.Info("Graceful shutdown completed")
}

// GetMetricsRegistry returns the Prometheus metrics registry
func (ac *AdxServerContext) GetMetricsRegistry() *prometheus.Registry {
	return ac.MetricsRegistry
}

// GetHTTPClient returns the HTTP client for external API calls
func (ac *AdxServerContext) GetHTTPClient() *http.Client {
	return ac.HTTPClient
}

// GetLogger returns the application logger
func (ac *AdxServerContext) GetLogger() logger.Logger {
	return ac.Logger
}

// GetSSPFactory returns the SSP adapter factory
// 获取SSP适配器工厂
func (ac *AdxServerContext) GetSSPFactory() *sspadapter.SSPAdapterFactory {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.SSPFactory
}

// SetSSPFactory sets the SSP adapter factory
// 设置SSP适配器工厂
func (ac *AdxServerContext) SetSSPFactory(factory *sspadapter.SSPAdapterFactory) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.SSPFactory = factory
}
