package adxserver

import (
	"context"
	"fmt"

	"github.com/echoface/admux/internal/adx_engine/adxcore"
	"github.com/echoface/admux/internal/adx_engine/config"
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

// GetSSPAdapter retrieves SSP adapter and configuration based on context
// 根据上下文获取SSP适配器和配置
func (s *AdxServer) GetSSPAdapter(ctx context.Context) (adxcore.ISSPAdapter, *config.SSPConfig, error) {
	// Extract SSP ID from context
	sspID, ok := ctx.Value("sspid").(string)
	if !ok || sspID == "" {
		return nil, nil, fmt.Errorf("SSP ID not found in request context")
	}

	// Get SSP factory from application context
	sspFactory := s.appCtx.GetSSPFactory()
	if sspFactory == nil {
		return nil, nil, fmt.Errorf("SSP factory not initialized")
	}

	// Get adapter and configuration
	adapter, sspConfig, err := sspFactory.GetAdapter(sspID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get SSP adapter for ID %s: %v", sspID, err)
	}

	return adapter, sspConfig, nil
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
