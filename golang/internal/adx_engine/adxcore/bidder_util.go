package adxcore

import (
	"fmt"
)

// Convenience functions for working with the global bidder factory

// RegisterBidder registers a bidder with the global factory
func RegisterBidder(bidder Bidder) error {
	return GetGlobalBidderFactory().RegisterBidder(bidder)
}

// GetBidder retrieves a bidder from the global factory
func GetBidder(bidderID string) (Bidder, error) {
	return GetGlobalBidderFactory().GetBidder(bidderID)
}

// GetAllBidders returns all bidders from the global factory
func GetAllBidders() map[string]Bidder {
	return GetGlobalBidderFactory().GetAllBidders()
}

// HasBidder checks if a bidder exists in the global factory
func HasBidder(bidderID string) bool {
	return GetGlobalBidderFactory().HasBidder(bidderID)
}

// UnregisterBidder removes a bidder from the global factory
func UnregisterBidder(bidderID string) error {
	return GetGlobalBidderFactory().UnregisterBidder(bidderID)
}

// BidderCount returns the number of registered bidders in the global factory
func BidderCount() int {
	return GetGlobalBidderFactory().BidderCount()
}

// RegisterBidders registers multiple bidders at once
func RegisterBidders(bidders ...Bidder) error {
	factory := GetGlobalBidderFactory()
	for _, bidder := range bidders {
		if err := factory.RegisterBidder(bidder); err != nil {
			return fmt.Errorf("failed to register bidder %s: %w", bidder.GetInfo().ID, err)
		}
	}
	return nil
}
