package adxcore

import ()

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
)
