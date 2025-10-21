package adxcore

import "context"

// BidFeatureHan

// SSIDValidationStage validates SSID parameter in the bid request context
type SSIDValidationStage struct{}

func (s *SSIDValidationStage) Process(ctx *BidRequestCtx) error {
	ssid, ok := ctx.Value("ssid").(string)
	if !ok || ssid == "" {
		return ErrMissingSSID
	}
	return nil
}

// BidProcessingStage handles the core bid processing logic
type BidProcessingStage struct {
	Broadcaster Broadcaster
}

func (s *BidProcessingStage) Process(ctx *BidRequestCtx) error {
	// Broadcast to bidders
	responses, err := s.Broadcaster.Broadcast(ctx)
	if err != nil {
		return err
	}

	// Store responses in context
	ctx.SetCandidates(responses)

	return nil
}

// Broadcaster interface for sending bid requests to DSPs
type Broadcaster interface {
	Broadcast(ctx *BidRequestCtx) ([]*BidCandidate, error)
}

// SetCandidates sets the bid candidates in the context
func (ctx *BidRequestCtx) SetCandidates(candidates []*BidCandidate) {
	ctx.candidates = candidates
}

// GetCandidates returns the bid candidates from the context
func (ctx *BidRequestCtx) GetCandidates() []*BidCandidate {
	return ctx.candidates
}

// AddCandidate adds a single bid candidate to the context
func (ctx *BidRequestCtx) AddCandidate(candidate *BidCandidate) {
	ctx.candidates = append(ctx.candidates, candidate)
}

// ClearCandidates removes all candidates from the context
func (ctx *BidRequestCtx) ClearCandidates() {
	ctx.candidates = nil
}

// GetSSID safely extracts SSID from context
func (ctx *BidRequestCtx) GetSSID() (string, bool) {
	ssid, ok := ctx.Value("ssid").(string)
	return ssid, ok
}

// WithSSID creates a new context with SSID value
func WithSSID(parent context.Context, ssid string) context.Context {
	return context.WithValue(parent, "ssid", ssid)
}
