package adxserver

import (
	"context"
	"io"

	"github.com/echoface/admux/internal/ssadapter"
)

type PipelineStage interface {
	Process(ctx context.Context, data any) (any, error)
}

type AdxServer struct {
	stages []PipelineStage
}

func NewAdxServer() *AdxServer {
	return &AdxServer{
		stages: []PipelineStage{
			&SSIDValidationStage{},
			&BidProcessingStage{},
			// Add more stages as needed
		},
	}
}

func (s *AdxServer) ProcessBid(ctx context.Context, body io.Reader) (interface{}, error) {
	var currentData interface{} = body

	for _, stage := range s.stages {
		result, err := stage.Process(ctx, currentData)
		if err != nil {
			return nil, err
		}
		currentData = result
	}

	return currentData, nil
}

type SSIDValidationStage struct{}

func (s *SSIDValidationStage) Process(ctx context.Context, data interface{}) (interface{}, error) {
	ssid, ok := ctx.Value("ssid").(string)
	if !ok || ssid == "" {
		return nil, ErrMissingSSID
	}

	// Use SS adapter to validate and process
	adapter := ssadapter.NewSSAdapter(ssid)
	validatedCtx, err := adapter.Validate(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"ctx":  validatedCtx,
		"data": data,
		"ssid": ssid,
	}, nil
}

type BidProcessingStage struct{}

func (s *BidProcessingStage) Process(ctx context.Context, data interface{}) (interface{}, error) {
	// For now, return a stub response
	// TODO: Implement actual bid processing logic
	return map[string]interface{}{
		"id": "test-bid-id",
		"seatbid": []map[string]interface{}{
			{
				"bid": []map[string]interface{}{
					{
						"id":    "bid-1",
						"impid": "imp-1",
						"price": 1.23,
						"adm":   "<div>Test Ad</div>",
					},
				},
			},
		},
	}, nil
}

var ErrMissingSSID = &AdxError{Message: "missing ssid parameter", Code: "MISSING_SSID"}

type AdxError struct {
	Message string
	Code    string
}

func (e *AdxError) Error() string {
	return e.Message
}

