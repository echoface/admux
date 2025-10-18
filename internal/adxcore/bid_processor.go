package adxcore

import (
	"context"
	"fmt"

	"github.com/echoface/admux/api/gen/admux/openrtb"
	"google.golang.org/protobuf/proto"
)

// BidProcessor handles core bid processing logic
type BidProcessor struct {
	stages      []PipelineStage
	broadcaster Broadcaster
}

// NewBidProcessor creates a new bid processor with default stages
func NewBidProcessor(broadcaster Broadcaster) *BidProcessor {
	return &BidProcessor{
		stages: []PipelineStage{
			&SSIDValidationStage{},
			&BidProcessingStage{Broadcaster: broadcaster},
		},
		broadcaster: broadcaster,
	}
}

// ProcessBid processes a bid request through all stages
func (bp *BidProcessor) ProcessBid(ctx context.Context, bidReq *openrtb.BidRequest) (*BidRequestCtx, error) {
	// Create bid request context
	bidCtx := &BidRequestCtx{
		Context: ctx,
		Request: bidReq,
	}

	// Process through all stages
	for _, stage := range bp.stages {
		if err := stage.Process(bidCtx); err != nil {
			return nil, fmt.Errorf("pipeline stage failed: %w", err)
		}
	}

	return bidCtx, nil
}

// AddStage adds a custom processing stage
func (bp *BidProcessor) AddStage(stage PipelineStage) {
	bp.stages = append(bp.stages, stage)
}

// SelectWinner selects the winning bid from candidates (placeholder implementation)
func (bp *BidProcessor) SelectWinner(candidates []*BidCandidate) *BidCandidate {
	if len(candidates) == 0 {
		return nil
	}

	// Simple implementation: return first candidate
	// In a real implementation, this would apply auction logic
	return candidates[0]
}

// BuildBidResponse builds the final bid response (placeholder implementation)
func (bp *BidProcessor) BuildBidResponse(ctx *BidRequestCtx, winner *BidCandidate) (*openrtb.BidResponse, error) {
	if winner == nil {
		// Return empty bid response with no bids
		return &openrtb.BidResponse{
			Id:      proto.String(ctx.Request.GetId()),
			Seatbid: []*openrtb.BidResponse_SeatBid{},
		}, nil
	}

	// Clone the winning response
	response := &openrtb.BidResponse{
		Id:      proto.String(ctx.Request.GetId()),
		Seatbid: winner.Response.GetSeatbid(),
		Cur:     proto.String(winner.Response.GetCur()),
	}

	return response, nil
}
