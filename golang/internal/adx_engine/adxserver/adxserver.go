package adxserver

import (
	"context"

	"github.com/echoface/admux/internal/adx_engine/adxcore"
	admux_rtb "github.com/echoface/admux/pkg/protogen/admux"
)

type AdxServer struct {
	stages []adxcore.PipelineStage
	appCtx *AdxServerContext
}

func NewAdxServer(appCtx *AdxServerContext) *AdxServer {
	return &AdxServer{
		stages: []adxcore.PipelineStage{
			&adxcore.SSIDValidationStage{},
			&adxcore.BidProcessingStage{Broadcaster: NewBroadcastManager(appCtx)},
		},
		appCtx: appCtx,
	}
}

func (s *AdxServer) ProcessBid(ctx context.Context, bidReq *admux_rtb.BidRequest) (any, error) {
	// Create bid request context
	bidCtx := &adxcore.BidRequestCtx{
		Context: ctx,
		Request: bidReq,
	}

	// Process through pipeline stages
	for _, stage := range s.stages {
		err := stage.Process(bidCtx)
		if err != nil {
			return nil, err
		}
	}

	// Return processing results
	return s.buildResponse(bidCtx), nil
}

// buildResponse constructs the final response from processed bid context
func (s *AdxServer) buildResponse(bidCtx *adxcore.BidRequestCtx) map[string]any {
	// TODO: Implement response building logic
	// This should create the OpenRTB response from the winning bid
	return map[string]any{
		"status":           "processed",
		"candidates_count": len(bidCtx.GetCandidates()),
	}
}
