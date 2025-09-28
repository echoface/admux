package sspadapter

import (
	"context"
	"fmt"
)

type SSPAdapter interface {
	Validate(ctx context.Context) (context.Context, error)
	Process(ctx context.Context, data interface{}) (interface{}, error)
}

type SSPAdapterFactory struct{}

func NewSSPAdapter(ssid string) SSPAdapter {
	// For now, return a stub adapter
	// In real implementation, this would return different adapters based on ssid
	return &StubSSPAdapter{ssid: ssid}
}

type StubSSPAdapter struct {
	ssid string
}

func (a *StubSSPAdapter) Validate(ctx context.Context) (context.Context, error) {
	// Stub validation - in real implementation, this would validate the SSID
	// against a configuration or database
	if a.ssid == "" {
		return nil, fmt.Errorf("invalid ssid")
	}

	// Add SSID-specific context values
	ctx = context.WithValue(ctx, "validated_ssid", a.ssid)
	ctx = context.WithValue(ctx, "adapter_type", "stub")
	ctx = context.WithValue(ctx, "ssid_config", map[string]interface{}{
		"timeout_ms":     100,
		"max_bid_price":  5.0,
		"allowed_formats": []string{"banner", "video"},
	})

	return ctx, nil
}

func (a *StubSSPAdapter) Process(ctx context.Context, data interface{}) (interface{}, error) {
	// Stub processing - in real implementation, this would process the bid request
	// according to SSID-specific rules
	return map[string]interface{}{
		"status":    "processed",
		"ssid":      a.ssid,
		"timestamp": "2024-01-01T00:00:00Z",
		"data":      data,
	}, nil
}