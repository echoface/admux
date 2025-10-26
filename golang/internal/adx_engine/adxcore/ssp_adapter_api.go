package adxcore

import (
	"context"
	"time"

	admux_rtb "github.com/echoface/admux/pkg/protogen/admux"
	"github.com/echoface/admux/internal/adx_engine/config"
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

		// SSP相关信息
		SSPID      string                 // SSP标识符
		SSPConfig  *config.SSPConfig      // SSP配置信息

		// 竞价处理状态
		ProcessingStartTime time.Time     // 处理开始时间
		ProcessingStages    []string      // 已处理的管道阶段
		ProcessingErrors    []error       // 处理过程中的错误

		// 竞价候选者
		Candidates []*BidCandidate        // 所有DSP竞价响应
		FilteredCandidates []*BidCandidate // 过滤后的竞价候选者

		// 响应构建
		Response    *admux_rtb.BidResponse // 最终响应
		ResponseTime time.Time            // 响应时间
	}

	// BidCandidate represents a bid response from DSP
	BidCandidate struct {
		Response *admux_rtb.BidResponse
	}
)

// NewBidRequestCtx creates a new bid request context with initialized fields
func NewBidRequestCtx(parent context.Context, request *admux_rtb.BidRequest) *BidRequestCtx {
	return &BidRequestCtx{
		Context:              parent,
		Request:              request,
		ProcessingStartTime:  time.Now(),
		ProcessingStages:     make([]string, 0),
		ProcessingErrors:     make([]error, 0),
		Candidates:           make([]*BidCandidate, 0),
		FilteredCandidates:   make([]*BidCandidate, 0),
	}
}

// AddProcessingStage adds a processing stage to the context
func (ctx *BidRequestCtx) AddProcessingStage(stageName string) {
	ctx.ProcessingStages = append(ctx.ProcessingStages, stageName)
}

// AddProcessingError adds an error to the processing errors list
func (ctx *BidRequestCtx) AddProcessingError(err error) {
	ctx.ProcessingErrors = append(ctx.ProcessingErrors, err)
}

// SetResponse sets the final response and response time
func (ctx *BidRequestCtx) SetResponse(response *admux_rtb.BidResponse) {
	ctx.Response = response
	ctx.ResponseTime = time.Now()
}

// SetSSPInfo sets SSP-related information
func (ctx *BidRequestCtx) SetSSPInfo(sspID string, sspConfig *config.SSPConfig) {
	ctx.SSPID = sspID
	ctx.SSPConfig = sspConfig
}
