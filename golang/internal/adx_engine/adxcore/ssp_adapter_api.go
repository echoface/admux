package adxcore

import (
	"context"

	admux_rtb "github.com/echoface/admux/pkg/protogen/admux"
)

type (
	// Supply Side Adapter interface
	ISSPAdapter interface {
		ToInternalBidRequest(ctx *BidRequestCtx, data []byte) error

		PackSSPResponse(ctx *BidRequestCtx) ([]byte, error)
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

		Request *admux_rtb.BidRequest // 由ssp adapter 提供

		candidates []*BidCandidate // Fixed typo: was "canidates"
	}

	// BidCandidate represents a bid response from DSP
	BidCandidate struct {
		Response *admux_rtb.BidResponse
	}
)
