package dspbidder

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/echoface/admux/internal/adxserver"
)

// BaseBidder 基础DSP bidder实现
type BaseBidder struct {
	BidderID   string
	Endpoint   string
	QPSLimit   int
	Healthy    bool
	Timeout    time.Duration
}

// NewBaseBidder 创建基础bidder
func NewBaseBidder(bidderID, endpoint string, qpsLimit int, timeout time.Duration) *BaseBidder {
	return &BaseBidder{
		BidderID:  bidderID,
		Endpoint:  endpoint,
		QPSLimit:  qpsLimit,
		Healthy:  true,
		Timeout:   timeout,
	}
}

// GetBidderID 返回bidder标识符
func (b *BaseBidder) GetBidderID() string {
	return b.BidderID
}

// GetEndpoint 返回bidder端点URL
func (b *BaseBidder) GetEndpoint() string {
	return b.Endpoint
}

// GetQPSLimit 返回QPS限制
func (b *BaseBidder) GetQPSLimit() int {
	return b.QPSLimit
}

// SendBidRequest 发送竞价请求到DSP
func (b *BaseBidder) SendBidRequest(ctx context.Context, bidRequest any) (any, error) {
	// 模拟DSP响应延迟
	sleepTime := time.Duration(rand.Intn(100)) * time.Millisecond
	time.Sleep(sleepTime)

	// 模拟竞价响应
	return map[string]any{
		"bidder_id": b.BidderID,
		"bid": map[string]any{
			"id":    fmt.Sprintf("bid-%s-%d", b.BidderID, time.Now().UnixNano()),
			"impid": "imp-1",
			"price": rand.Float64() * 10, // 随机价格0-10
			"adm":   fmt.Sprintf("<div>Ad from %s</div>", b.BidderID),
		},
		"latency_ms": sleepTime.Milliseconds(),
	}, nil
}

// IsHealthy 检查bidder是否健康
func (b *BaseBidder) IsHealthy() bool {
	return b.Healthy
}

// SetHealthStatus 设置健康状态
func (b *BaseBidder) SetHealthStatus(healthy bool) {
	b.Healthy = healthy
}