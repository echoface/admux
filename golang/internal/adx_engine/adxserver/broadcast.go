package adxserver

import (
	"context"
	"time"

	"github.com/echoface/admux/internal/adx_engine/adxcore"
	"github.com/echoface/admux/internal/adx_engine/health"
	"github.com/echoface/admux/internal/adx_engine/metrics"
	"github.com/echoface/admux/pkg/concurrent"
	"github.com/echoface/admux/pkg/retry"
)

// BidResponse 竞价响应结构
type BidResponse struct {
	BidderID   string
	Candidates []*adxcore.BidCandidate

	Error   error
	Latency time.Duration
}

// BroadcastManager 广播管理器
type BroadcastManager struct {
	appCtx         *AdxServerContext
	healthChecker  *health.HealthChecker
	circuitBreaker *health.CircuitBreaker
	metrics        *metrics.BroadcastMetrics
}

// NewBroadcastManager 创建广播管理器
func NewBroadcastManager(appCtx *AdxServerContext) *BroadcastManager {
	return &BroadcastManager{
		appCtx:         appCtx,
		healthChecker:  health.NewHealthChecker(30*time.Second, 5, 3),
		circuitBreaker: health.NewCircuitBreaker(),
		metrics:        metrics.NewBroadcastMetrics("adx", "broadcast"),
	}
}

// BroadcastToBidders 向定向的DSP bidder广播竞价请求
func (bm *BroadcastManager) BroadcastToBidders(bidRequest *adxcore.BidRequestCtx, bidders []adxcore.Bidder) ([]BidResponse, error) {
	// 记录请求指标
	bm.metrics.RecordRequest()

	// 过滤健康的bidder
	healthyBidders := bm.filterHealthyBidders(bidders)
	if len(healthyBidders) == 0 {
		return nil, nil
	}

	// 更新活跃bidder指标
	bm.metrics.ActiveBidders.Set(float64(len(healthyBidders)))

	// 创建并发控制器
	controller := concurrent.NewConcurrencyController(bm.appCtx.Config.MaxConnections)

	// 准备任务列表
	tasks := make([]concurrent.Task[BidResponse], 0, len(healthyBidders))
	for _, bidder := range healthyBidders {
		bidder := bidder // 创建局部变量
		tasks = append(tasks, func(ctx context.Context) (BidResponse, error) {
			return bm.executeBidderRequest(ctx, bidder, bidRequest)
		})
	}

	// 执行并发任务，使用SSP级别的超时时间
	sspTimeout := bm.getSSPTimeout(bidRequest)
	results, err := concurrent.ExecuteWithTimeout(controller, bidRequest.Context, tasks, sspTimeout)
	if err != nil {
		return nil, err
	}

	// 转换结果
	responses := make([]BidResponse, 0, len(results))
	for _, result := range results {
		responses = append(responses, result.Value)
	}

	return responses, nil
}

// BroadcastWithBidders 向指定的bidder列表广播竞价请求
func (bm *BroadcastManager) BroadcastWithBidders(ctx *adxcore.BidRequestCtx, bidders []adxcore.Bidder) ([]*adxcore.BidCandidate, error) {
	responses, err := bm.BroadcastToBidders(ctx, bidders)
	if err != nil {
		return nil, err
	}

	// Convert BidResponse to BidCandidate and collect all candidates
	var allCandidates []*adxcore.BidCandidate
	successCount := 0
	failureCount := 0

	for _, response := range responses {
		// 记录响应指标
		success := response.Error == nil
		latencySeconds := response.Latency.Seconds()
		bm.metrics.RecordResponse(success, latencySeconds)

		if success {
			successCount++
			allCandidates = append(allCandidates, response.Candidates...)
		} else {
			failureCount++
		}

		// 记录健康状态
		bm.healthChecker.UpdateHealthStatus(response.BidderID, success, response.Error)
	}

	// 更新健康状态指标
	bm.updateHealthMetrics()

	return allCandidates, nil
}

// filterHealthyBidders 过滤健康的bidder列表
func (bm *BroadcastManager) filterHealthyBidders(bidders []adxcore.Bidder) []adxcore.Bidder {
	healthyBidders := make([]adxcore.Bidder, 0, len(bidders))
	for _, bidder := range bidders {
		bidderID := bidder.GetInfo().ID
		if bm.healthChecker.IsHealthy(bidderID) && bm.circuitBreaker.Allow() {
			healthyBidders = append(healthyBidders, bidder)
		}
	}

	return healthyBidders
}

// executeBidderRequest 执行单个bidder请求
func (bm *BroadcastManager) executeBidderRequest(
	ctx context.Context,
	bidder adxcore.Bidder,
	bidRequest *adxcore.BidRequestCtx,
) (BidResponse, error) {
	response := BidResponse{
		BidderID: bidder.GetInfo().ID,
	}

	startTime := time.Now()

	// 创建带超时的上下文
	bidderTimeout := bm.getBidderTimeout(bidder)
	bidderCtx, cancel := context.WithTimeout(ctx, bidderTimeout)
	defer cancel()

	// 使用重试机制执行请求
	retryConfig := &retry.RetryConfig{
		MaxRetries:        2,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		BackoffMultiplier: 2.0,
	}

	candidates, err := retry.Retry(bidderCtx, func(ctx context.Context) ([]*adxcore.BidCandidate, error) {
		return bidder.SendBidRequest(bidRequest)
	}, retryConfig)

	response.Latency = time.Since(startTime)

	if err != nil {
		response.Error = err
		bm.circuitBreaker.RecordFailure()
	} else {
		response.Candidates = candidates
		bm.circuitBreaker.RecordSuccess()
	}

	return response, nil
}

// getSSPTimeout 获取SSP级别的超时时间
func (bm *BroadcastManager) getSSPTimeout(bidRequest *adxcore.BidRequestCtx) time.Duration {
	if bidRequest.SSPConfig != nil && bidRequest.SSPConfig.Timeout > 0 {
		return bidRequest.SSPConfig.Timeout
	}
	return 3000 * time.Millisecond // 默认SSP超时
}

// getBidderTimeout 获取bidder级别的超时时间
func (bm *BroadcastManager) getBidderTimeout(bidder adxcore.Bidder) time.Duration {
	// 这里可以根据bidder配置获取超时时间
	// 目前返回默认值
	return 2000 * time.Millisecond // 默认bidder超时
}

// updateHealthMetrics 更新健康状态指标
func (bm *BroadcastManager) updateHealthMetrics() {
	allStatus := bm.healthChecker.GetAllHealthStatus()

	healthyCount := 0
	unhealthyCount := 0

	for _, status := range allStatus {
		if status.Healthy {
			healthyCount++
		} else {
			unhealthyCount++
		}
	}

	bm.metrics.UpdateBidderHealth(healthyCount, unhealthyCount)
}
