package adxcore

import (
	"context"

	"github.com/echoface/admux/api/gen/admux/openrtb"
)

type (
	// Supply Side Adapter interface
	ISSAdapter interface {
		BuildBidRequest(ctx *BidRequestCtx, data []byte) error

		PackResponse(ctx *BidRequestCtx)
	}

	ServerContext struct {
		// 维护引擎上下文内容，或者是全局的逻辑
	}

	BidRequestCtx struct {
		context.Context

		Request *openrtb.BidRequest

		canidates []*BidCandidate
	}

	BidCandidate struct {
		Response *openrtb.BidResponse
	}
)
