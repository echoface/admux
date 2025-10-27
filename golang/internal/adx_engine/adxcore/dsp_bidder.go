package adxcore

import (
	"fmt"
	"sync"
)

type BidderInfo struct {
	ID string `json:"id,omitempty" yaml:"id"`

	QPS      int    `json:"qps,omitempty" yaml:"qps"`
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint"`
}

// Bidder defines the interface that all DSP bidders must implement
type Bidder interface {
	// GetBidderID returns the unique identifier for this bidder
	GetInfo() *BidderInfo

	// SendBidRequest sends a bid request to the DSP and returns bid candidates
	SendBidRequest(bidRequest *BidRequestCtx) ([]*BidCandidate, error)
}

// BidderFactory manages the registration and retrieval of DSP bidders
type BidderFactory struct {
	bidders map[string]Bidder
	mu      sync.RWMutex
}

// NewBidderFactory creates a new bidder factory
func NewBidderFactory() *BidderFactory {
	return &BidderFactory{
		bidders: make(map[string]Bidder),
	}
}

// RegisterBidder registers a bidder with the factory
func (f *BidderFactory) RegisterBidder(bidder Bidder) error {
	if bidder == nil {
		return fmt.Errorf("cannot register nil bidder")
	}
	bidderID := bidder.GetInfo().ID
	if bidderID == "" {
		return fmt.Errorf("bidder ID cannot be empty")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.bidders[bidderID]; exists {
		return fmt.Errorf("bidder with ID '%s' is already registered", bidderID)
	}

	f.bidders[bidderID] = bidder
	return nil
}

// GetBidder retrieves a bidder by its ID
func (f *BidderFactory) GetBidder(bidderID string) (Bidder, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	bidder, exists := f.bidders[bidderID]
	if !exists {
		return nil, fmt.Errorf("bidder with ID '%s' not found", bidderID)
	}

	return bidder, nil
}

// GetAllBidders returns all registered bidders
func (f *BidderFactory) GetAllBidders() map[string]Bidder {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Return a copy to avoid external modification
	biddersCopy := make(map[string]Bidder)
	for id, bidder := range f.bidders {
		biddersCopy[id] = bidder
	}

	return biddersCopy
}

// UnregisterBidder removes a bidder from the factory
func (f *BidderFactory) UnregisterBidder(bidderID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.bidders[bidderID]; !exists {
		return fmt.Errorf("bidder with ID '%s' not found", bidderID)
	}

	delete(f.bidders, bidderID)
	return nil
}

// HasBidder checks if a bidder with the given ID is registered
func (f *BidderFactory) HasBidder(bidderID string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	_, exists := f.bidders[bidderID]
	return exists
}

// BidderCount returns the number of registered bidders
func (f *BidderFactory) BidderCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return len(f.bidders)
}

// Global bidder factory instance
var (
	globalBidderFactory *BidderFactory
	factoryOnce         sync.Once
)

// GetGlobalBidderFactory returns the global bidder factory instance (singleton)
func GetGlobalBidderFactory() *BidderFactory {
	factoryOnce.Do(func() {
		globalBidderFactory = NewBidderFactory()
	})
	return globalBidderFactory
}
