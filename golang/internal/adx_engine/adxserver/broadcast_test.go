package adxserver

import (
	"context"
	"testing"
	"time"

	"github.com/echoface/admux/internal/adx_engine/adxcore"
	"github.com/echoface/admux/internal/adx_engine/config"
	admux_rtb "github.com/echoface/admux/pkg/protogen/admux"
	"github.com/stretchr/testify/assert"
)

// MockBidder 模拟bidder实现
type MockBidder struct {
	id       string
	healthy  bool
	latency  time.Duration
	shouldError bool
}

func NewMockBidder(id string, healthy bool, latency time.Duration, shouldError bool) *MockBidder {
	return &MockBidder{
		id:          id,
		healthy:     healthy,
		latency:     latency,
		shouldError: shouldError,
	}
}

func (m *MockBidder) GetInfo() *adxcore.BidderInfo {
	return &adxcore.BidderInfo{
		ID:       m.id,
		QPS:      100,
		Endpoint: "http://mock-dsp.com",
	}
}

func (m *MockBidder) SendBidRequest(bidRequest *adxcore.BidRequestCtx) ([]*adxcore.BidCandidate, error) {
	if m.shouldError {
		return nil, &retry.RetryableError{
			Type:    retry.NetworkError,
			Message: "mock network error",
		}
	}

	// 模拟延迟
	time.Sleep(m.latency)

	// 返回模拟竞价候选
	candidate := &adxcore.BidCandidate{
		Response: &admux_rtb.BidResponse{
			Id: "mock-bid",
		},
		CPMPrice: 500,
	}

	return []*adxcore.BidCandidate{candidate}, nil
}

func TestBroadcastManager_BroadcastToBidders(t *testing.T) {
	// 创建测试上下文
	ctx := context.Background()

	// 创建模拟服务器上下文
	appCtx := &AdxServerContext{
		Config: &config.ServerConfig{
			MaxConnections: 10,
		},
	}

	// 创建广播管理器
	bm := NewBroadcastManager(appCtx)

	// 注册模拟bidder
	factory := adxcore.GetGlobalBidderFactory()

	// 注册健康的bidder
	healthyBidder := NewMockBidder("healthy-1", true, 50*time.Millisecond, false)
	factory.RegisterBidder(healthyBidder)

	// 注册会失败的bidder
	failingBidder := NewMockBidder("failing-1", true, 10*time.Millisecond, true)
	factory.RegisterBidder(failingBidder)

	// 创建竞价请求上下文
	bidRequest := adxcore.NewBidRequestCtx(ctx, &admux_rtb.BidRequest{
		Id: "test-request",
	})

	// 执行广播
	responses, err := bm.BroadcastToBidders(bidRequest)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, responses)
	assert.Greater(t, len(responses), 0)

	// 验证响应
	var successCount, failureCount int
	for _, response := range responses {
		if response.Error == nil {
			successCount++
			assert.NotEmpty(t, response.BidderID)
			assert.NotNil(t, response.Candidates)
			assert.Greater(t, response.Latency, time.Duration(0))
		} else {
			failureCount++
			assert.NotEmpty(t, response.BidderID)
			assert.Error(t, response.Error)
		}
	}

	// 应该至少有一个成功和一个失败
	assert.Greater(t, successCount, 0)
	assert.Greater(t, failureCount, 0)
}

func TestBroadcastManager_GetHealthyBidders(t *testing.T) {
	// 创建测试上下文
	appCtx := &AdxServerContext{
		Config: &config.ServerConfig{
			MaxConnections: 10,
		},
	}

	// 创建广播管理器
	bm := NewBroadcastManager(appCtx)

	// 注册模拟bidder
	factory := adxcore.GetGlobalBidderFactory()

	// 注册健康的bidder
	healthyBidder1 := NewMockBidder("healthy-1", true, 10*time.Millisecond, false)
	healthyBidder2 := NewMockBidder("healthy-2", true, 10*time.Millisecond, false)
	factory.RegisterBidder(healthyBidder1)
	factory.RegisterBidder(healthyBidder2)

	// 获取健康bidder
	healthyBidders := bm.getHealthyBidders()

	// 验证结果
	assert.Equal(t, 2, len(healthyBidders))

	// 验证bidder ID
	bidderIDs := make(map[string]bool)
	for _, bidder := range healthyBidders {
		bidderIDs[bidder.GetInfo().ID] = true
	}

	assert.True(t, bidderIDs["healthy-1"])
	assert.True(t, bidderIDs["healthy-2"])
}

func TestBroadcastManager_TimeoutHandling(t *testing.T) {
	// 创建测试上下文
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 创建模拟服务器上下文
	appCtx := &AdxServerContext{
		Config: &config.ServerConfig{
			MaxConnections: 10,
		},
	}

	// 创建广播管理器
	bm := NewBroadcastManager(appCtx)

	// 注册模拟bidder（有高延迟）
	factory := adxcore.GetGlobalBidderFactory()
	slowBidder := NewMockBidder("slow-1", true, 200*time.Millisecond, false)
	factory.RegisterBidder(slowBidder)

	// 创建竞价请求上下文
	bidRequest := adxcore.NewBidRequestCtx(ctx, &admux_rtb.BidRequest{
		Id: "test-timeout",
	})

	// 执行广播（应该超时）
	responses, err := bm.BroadcastToBidders(bidRequest)

	// 验证超时处理
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deadline exceeded")
	assert.Nil(t, responses)
}