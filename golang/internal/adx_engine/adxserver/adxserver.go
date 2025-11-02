package adxserver

import (
	"fmt"

	"github.com/bytedance/gg/gslice"

	"github.com/echoface/admux/internal/adx_engine/adxcore"
	"github.com/echoface/admux/internal/adx_engine/config"
)

type AdxServer struct {
	appCtx *AdxServerContext
}

func NewAdxServer(appCtx *AdxServerContext) *AdxServer {
	return &AdxServer{
		appCtx: appCtx,
	}
}

func (s *AdxServer) ProcessBid(bixCtx *adxcore.BidRequestCtx) (err error) {
	// Create bid request context

	// 1. 特征补全
	if err := s.completeFeatures(bixCtx); err != nil {
		return fmt.Errorf("completeFeatures fail:%w", err)
	}

	// 2. DSP targeting 定向，找出本次流量需要请求哪些DSP需要竞价调用
	var bidders []adxcore.Bidder
	if bidders, err = s.targetingBidders(bixCtx); err != nil {
		return fmt.Errorf("targetingBidders fail:%w", err)
	}

	// 3. DSP竞价广播
	if err := s.broadcast(bixCtx, bidders); err != nil {
		return fmt.Errorf("broadcast fail:%w", err)
	}

	// 4. 素材Collector
	if err := s.broadcast(bixCtx, bidders); err != nil {
		return fmt.Errorf("broadcast fail:%w", err)
	}

	// 5. 候选Filters
	// 串并过滤支持；编排后执行
	if err := s.filterCanidates(bixCtx); err != nil {
		return fmt.Errorf("filterCanidates fail:%w", err)
	}

	// 6. Ranking: 找出ECPM最高的竞价（最有可能获得最大利润）的候选广告
	// ecpm * winrate
	if err := s.rankingCandidates(bixCtx); err != nil {
		return fmt.Errorf("rankingCandidates fail:%s", err)
	}

	// Return processing results
	_ = s.buildResponse(bixCtx)
	return nil
}

func (s *AdxServer) completeFeatures(ctx *adxcore.BidRequestCtx) error {
	// 这里通过各种feature provider 补全特征数据

	return nil
}

// targetingBidders 根据各bidder 和 ssp的要求，得到最终需要广播的竞价方
func (s *AdxServer) targetingBidders(ctx *adxcore.BidRequestCtx) ([]adxcore.Bidder, error) {
	return make([]adxcore.Bidder, 0), nil
}

func (s *AdxServer) broadcast(ctx *adxcore.BidRequestCtx, bidders []adxcore.Bidder) error {
	// 创建广播管理器
	broadcastManager := NewBroadcastManager(s.appCtx)

	// 向定向的bidder广播竞价请求
	candidates, err := broadcastManager.BroadcastWithBidders(ctx, bidders)
	if err != nil {
		return fmt.Errorf("broadcast to bidders failed: %w", err)
	}

	// 将竞价候选者添加到上下文中
	for _, candidate := range candidates {
		ctx.AddCandidate(candidate)
	}

	return nil
}

func (s *AdxServer) filterCanidates(ctx *adxcore.BidRequestCtx) error {
	canidates := ctx.GetCandidates()
	ctx.Candidates = make([]*adxcore.BidCandidate, 0, len(canidates)/2)

	return nil
}

func (s *AdxServer) rankingCandidates(ctx *adxcore.BidRequestCtx) error {
	gslice.SortBy(ctx.GetCandidates(), func(l, r *adxcore.BidCandidate) bool {
		return l.CPMPrice < r.CPMPrice
	})
	return nil
}

// GetSSPAdapter retrieves SSP adapter and configuration based on context
// 根据上下文获取SSP适配器和配置
func (s *AdxServer) GetSSPAdapter(sspID string) (adxcore.ISSPAdapter, *config.SSPConfig, error) {
	// Extract SSP ID from context
	if sspID == "" {
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
