package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/echoface/admux/internal/adxserver"
	"github.com/echoface/admux/internal/config"
	"github.com/echoface/admux/internal/dspbidder"
)

func main() {
	// 初始化应用上下文
	cfg := config.NewDefaultConfig()
	appCtx := adxserver.NewAppContext(cfg)

	// 注册几个示例DSP bidder
	bidder1 := dspbidder.NewBaseBidder("dsp1", "https://dsp1.com/bid", 100, 2*time.Second)
	bidder2 := dspbidder.NewBaseBidder("dsp2", "https://dsp2.com/bid", 150, 2*time.Second)
	bidder3 := dspbidder.NewBaseBidder("dsp3", "https://dsp3.com/bid", 200, 2*time.Second)

	appCtx.RegisterDSPBidder(bidder1)
	appCtx.RegisterDSPBidder(bidder2)
	appCtx.RegisterDSPBidder(bidder3)

	// 创建ADX服务器
	adxServer := adxserver.NewAdxServer(appCtx)

	// 创建示例竞价请求
	bidRequest := map[string]any{
		"id": "test-request-1",
		"imp": []map[string]any{
			{
				"id":     "imp-1",
				"banner": map[string]any{"w": 300, "h": 250},
			},
		},
		"app": map[string]any{
			"id":  "test-app",
			"name": "Test Application",
		},
		"device": map[string]any{
			"ua": "Mozilla/5.0...",
			"ip": "192.168.1.1",
		},
	}

	// 处理竞价请求
	ctx := context.WithValue(context.Background(), "ssid", "test-ssp")

	fmt.Println("开始竞价广播...")
	startTime := time.Now()

	response, err := adxServer.ProcessBid(ctx, bidRequest)
	if err != nil {
		log.Fatalf("竞价处理失败: %v", err)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("竞价完成，耗时: %v\n", elapsed)
	fmt.Printf("最终响应: %+v\n", response)
}