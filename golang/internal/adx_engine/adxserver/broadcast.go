package adxserver

import (
	"sync"
	"time"

	"github.com/echoface/admux/internal/adx_engine/adxcore"
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
	appCtx *AdxServerContext
}

// NewBroadcastManager 创建广播管理器
func NewBroadcastManager(appCtx *AdxServerContext) *BroadcastManager {
	return &BroadcastManager{
		appCtx: appCtx,
	}
}

// BroadcastToBidders 向所有健康的DSP bidder广播竞价请求
func (bm *BroadcastManager) BroadcastToBidders(bidRequest *adxcore.BidRequestCtx) ([]BidResponse, error) {
	// 获取所有健康的bidder
	healthyBidders := bm.appCtx.GetHealthyDSPBidders()

	if len(healthyBidders) == 0 {
		return nil, nil
	}

	// 使用WaitGroup等待所有竞价完成
	var wg sync.WaitGroup
	responses := make([]BidResponse, 0, len(healthyBidders))
	responseChan := make(chan BidResponse, len(healthyBidders))

	// 为每个bidder启动goroutine
	for _, bidder := range healthyBidders {
		wg.Add(1)
		go func(bidder DSPBidder) {
			defer wg.Done()

			startTime := time.Now()
			bid, err := bidder.SendBidRequest(bidRequest)
			latency := time.Since(startTime)

			responseChan <- BidResponse{
				BidderID:   bidder.GetBidderID(),
				Candidates: bid,
				Error:      err,
				Latency:    latency,
			}
		}(bidder)
	}

	// 等待所有goroutine完成
	go func() {
		wg.Wait()
		close(responseChan)
	}()

	// 收集响应
	for response := range responseChan {
		responses = append(responses, response)
	}

	return responses, nil
}

// Broadcast implements adxcore.Broadcaster interface
func (bm *BroadcastManager) Broadcast(ctx *adxcore.BidRequestCtx) ([]*adxcore.BidCandidate, error) {
	responses, err := bm.BroadcastToBidders(ctx)
	if err != nil {
		return nil, err
	}

	// Convert BidResponse to BidCandidate and collect all candidates
	var allCandidates []*adxcore.BidCandidate
	for _, response := range responses {
		if response.Error == nil {
			allCandidates = append(allCandidates, response.Candidates...)
		}
		// TODO: Log errors appropriately instead of ignoring them
	}

	return allCandidates, nil
}
