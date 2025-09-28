package adxserver

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/echoface/admux/internal/adxcore"
	"github.com/echoface/admux/internal/config"
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

	// SSP adapters
	SSPAdapters map[string]SSPAdapter

	// DSP bidders
	DSPBidders map[string]DSPBidder

	// HTTP client for external API calls
	HTTPClient *http.Client

	// Context for graceful shutdown
	ShutdownCtx    context.Context
	ShutdownCancel context.CancelFunc

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Logger
	Logger *log.Logger

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

// DSPBidder interface for DSP integration
type DSPBidder interface {
	// GetBidderID returns the bidder identifier
	GetBidderID() string
	// GetEndpoint returns the bidder endpoint URL
	GetEndpoint() string
	// GetQPSLimit returns the QPS limit for this bidder
	GetQPSLimit() int
	// IsHealthy returns the health status of the bidder
	IsHealthy() bool
	// SendBidRequest sends bid request to DSP
	SendBidRequest(bidRequest *adxcore.BidRequestCtx) ([]*adxcore.BidCandidate, error)
}

// NewAppContext creates and initializes a new application context
func NewAppContext(cfg *config.ServerConfig) *AdxServerContext {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

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
		Addr:         cfg.Host + ":" + string(rune(cfg.Port)),
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
		DSPBidders:      make(map[string]DSPBidder),
		HTTPClient:      httpClient,
		ShutdownCtx:     shutdownCtx,
		ShutdownCancel:  shutdownCancel,
		Logger:          log.Default(),
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

// RegisterDSPBidder registers a new DSP bidder
func (ac *AdxServerContext) RegisterDSPBidder(bidder DSPBidder) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.DSPBidders[bidder.GetBidderID()] = bidder
}

// GetDSPBidder retrieves a DSP bidder by ID
func (ac *AdxServerContext) GetDSPBidder(bidderID string) (DSPBidder, bool) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	bidder, exists := ac.DSPBidders[bidderID]
	return bidder, exists
}

// GetAllDSPBidders returns all registered DSP bidders
func (ac *AdxServerContext) GetAllDSPBidders() []DSPBidder {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	bidders := make([]DSPBidder, 0, len(ac.DSPBidders))
	for _, bidder := range ac.DSPBidders {
		bidders = append(bidders, bidder)
	}
	return bidders
}

// GetHealthyDSPBidders returns all healthy DSP bidders
func (ac *AdxServerContext) GetHealthyDSPBidders() []DSPBidder {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	healthyBidders := make([]DSPBidder, 0)
	for _, bidder := range ac.DSPBidders {
		if bidder.IsHealthy() {
			healthyBidders = append(healthyBidders, bidder)
		}
	}
	return healthyBidders
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

	ac.Logger.Println("Initiating graceful shutdown...")
	ac.IsHealthy = false

	if ac.ShutdownCancel != nil {
		ac.ShutdownCancel()
	}

	// Close HTTP client connections
	if ac.HTTPClient != nil {
		ac.HTTPClient.CloseIdleConnections()
	}

	ac.Logger.Println("Graceful shutdown completed")
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
func (ac *AdxServerContext) GetLogger() *log.Logger {
	return ac.Logger
}
