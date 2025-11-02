package adxserver

import (
	"context"
	"testing"
	"time"

	"github.com/echoface/admux/internal/adx_engine/adxcore"
	"github.com/echoface/admux/internal/adx_engine/config"
	"github.com/echoface/admux/internal/adx_engine/dspbidder"
	admux_rtb "github.com/echoface/admux/pkg/protogen/admux"
	"github.com/stretchr/testify/assert"
)

// TestBroadcastIntegration 测试广播集成
func TestBroadcastIntegration(t *testing.T) {
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

	// 注册真实的DSP bidder
	factory := adxcore.GetGlobalBidderFactory()

	// 注册多个bidder
	bidder1 := dspbidder.NewBaseBidder("dsp-1", "http://dsp1.com", 100, 2000*time.Millisecond)
	bidder2 := dspbidder.NewBaseBidder("dsp-2", "http://dsp2.com", 100, 2000*time.Millisecond)
	bidder3 := dspbidder.NewBaseBidder("dsp-3", "http://dsp3.com", 100, 2000*time.Millisecond)

	factory.RegisterBidder(bidder1)
	factory.RegisterBidder(bidder2)
	factory.RegisterBidder(bidder3)

	// 创建竞价请求上下文
	bidRequest := adxcore.NewBidRequestCtx(ctx, &admux_rtb.BidRequest{
		Id: "integration-test",
	})

	// 执行广播
	startTime := time.Now()
	candidates, err := bm.Broadcast(bidRequest)
	elapsed := time.Since(startTime)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, candidates)

	// 验证性能：广播应该在合理时间内完成
	assert.Less(t, elapsed, 3*time.Second, "Broadcast should complete within 3 seconds")

	// 验证候选者数量
	assert.Greater(t, len(candidates), 0, "Should receive at least one candidate")

	// 验证候选者结构
	for _, candidate := range candidates {
		assert.NotNil(t, candidate)
		assert.Greater(t, candidate.CPMPrice, int64(0), "CPM price should be positive")
	}

	// 验证健康检查器状态
	healthStatus := bm.healthChecker.GetAllHealthStatus()
	assert.Equal(t, 3, len(healthStatus), "Should have health status for all 3 bidders")

	// 所有bidder应该都是健康的
	for bidderID, status := range healthStatus {
		assert.True(t, status.Healthy, "Bidder %s should be healthy", bidderID)
	}
}

// TestBroadcastConcurrent 测试并发广播
func TestBroadcastConcurrent(t *testing.T) {
	// 创建测试上下文
	ctx := context.Background()

	// 创建模拟服务器上下文
	appCtx := &AdxServerContext{
		Config: &config.ServerConfig{
			MaxConnections: 5, // 限制并发数
		},
	}

	// 创建广播管理器
	bm := NewBroadcastManager(appCtx)

	// 注册多个bidder
	factory := adxcore.GetGlobalBidderFactory()

	// 注册10个bidder来测试并发控制
	for i := 0; i < 10; i++ {
		bidder := dspbidder.NewBaseBidder(
			"concurrent-dsp-"+string(rune('a'+i)),
			"http://dsp-concurrent.com",
			100,
			1000*time.Millisecond,
		)
		factory.RegisterBidder(bidder)
	}

	// 创建竞价请求上下文
	bidRequest := adxcore.NewBidRequestCtx(ctx, &admux_rtb.BidRequest{
		Id: "concurrent-test",
	})

	// 执行广播
	startTime := time.Now()
	candidates, err := bm.Broadcast(bidRequest)
	elapsed := time.Since(startTime)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, candidates)

	// 验证并发控制：即使有10个bidder，也应该在合理时间内完成
	assert.Less(t, elapsed, 5*time.Second, "Concurrent broadcast should complete within 5 seconds")

	// 应该收到多个候选者
	assert.Greater(t, len(candidates), 5, "Should receive multiple candidates from concurrent bidders")
}

// TestBroadcastHealthCheck 测试健康检查功能
func TestBroadcastHealthCheck(t *testing.T) {
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

	// 注册bidder
	factory := adxcore.GetGlobalBidderFactory()

	// 注册健康的bidder
	healthyBidder := dspbidder.NewBaseBidder("healthy-dsp", "http://healthy-dsp.com", 100, 2000*time.Millisecond)
	factory.RegisterBidder(healthyBidder)

	// 创建竞价请求上下文
	bidRequest := adxcore.NewBidRequestCtx(ctx, &admux_rtb.BidRequest{
		Id: "health-check-test",
	})

	// 多次执行广播来测试健康检查
	for i := 0; i < 3; i++ {
		candidates, err := bm.Broadcast(bidRequest)
		assert.NoError(t, err)
		assert.NotNil(t, candidates)
	}

	// 验证健康状态
	healthStatus := bm.healthChecker.GetHealthStatus("healthy-dsp")
	assert.NotNil(t, healthStatus)
	assert.True(t, healthStatus.Healthy)
	assert.Greater(t, healthStatus.SuccessCount, 0)
	assert.Equal(t, 0, healthStatus.FailureCount)
}