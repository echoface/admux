package adxcore

import (
	"context"

	"github.com/echoface/admux/api/gen/admux/openrtb"
)

type (
	// Supply Side Adapter interface
	ISSPAdapter interface {
		BuildBidRequest(ctx *BidRequestCtx, data []byte) error

		PackResponse(ctx *BidRequestCtx)
	}

	// PipelineStage interface for bid request processing pipeline
	PipelineStage interface {
		Process(ctx *BidRequestCtx) error
	}

	// ServerContext struct
	ServerContext struct {
		// 维护引擎上下文内容，或者是全局的逻辑
	}

	// BidRequestCtx represents the context for bid request processing
	BidRequestCtx struct {
		context.Context

		Request    *openrtb.BidRequest
		candidates []*BidCandidate // Fixed typo: was "canidates"
	}

	// BidCandidate represents a bid response from DSP
	BidCandidate struct {
		Response *openrtb.BidResponse
	}

	// AdxError represents an ADX specific error
	AdxError struct {
		Message string
		Code    string
	}
)

// Error method for AdxError
func (e *AdxError) Error() string {
	return e.Message
}

// Predefined errors
var (
	ErrMissingSSID = &AdxError{Message: "missing ssid parameter", Code: "MISSING_SSID"}
)
