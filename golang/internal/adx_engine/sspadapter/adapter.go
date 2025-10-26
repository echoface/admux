package sspadapter

import (
	"encoding/json"
	"fmt"
	"sync"

	admux_rtb "github.com/echoface/admux/pkg/protogen/admux"
	"github.com/echoface/admux/internal/adx_engine/adxcore"
	"github.com/echoface/admux/internal/adx_engine/config"
	"github.com/echoface/admux/internal/adx_engine/sspadapter/kuaishou"
)

// SSPAdapterFactory manages SSP adapter instances
// 管理SSP适配器实例的工厂
type SSPAdapterFactory struct {
	mu       sync.RWMutex
	adapters map[string]adxcore.ISSPAdapter // SSP ID -> Adapter mapping
	configs  map[string]*config.SSPConfig   // SSP ID -> Config mapping
}

// NewSSPAdapterFactory creates a new SSP adapter factory
// 创建新的SSP适配器工厂
func NewSSPAdapterFactory(sspConfigs []config.SSPConfig) *SSPAdapterFactory {
	factory := &SSPAdapterFactory{
		adapters: make(map[string]adxcore.ISSPAdapter),
		configs:  make(map[string]*config.SSPConfig),
	}

	// Register all configured SSPs
	// 注册所有已配置的SSP
	for _, cfg := range sspConfigs {
		if cfg.Enabled {
			factory.configs[cfg.ID] = &cfg
			factory.adapters[cfg.ID] = factory.createAdapter(cfg.ID, cfg.Protocol)
		}
	}

	return factory
}

// GetAdapter returns the SSP adapter for the given SSP ID
// 根据SSP ID获取对应的适配器
func (f *SSPAdapterFactory) GetAdapter(sspID string) (adxcore.ISSPAdapter, *config.SSPConfig, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	adapter, exists := f.adapters[sspID]
	if !exists {
		return nil, nil, fmt.Errorf("SSP adapter not found for ID: %s", sspID)
	}

	config, exists := f.configs[sspID]
	if !exists {
		return nil, nil, fmt.Errorf("SSP config not found for ID: %s", sspID)
	}

	return adapter, config, nil
}

// createAdapter creates a new adapter instance based on protocol type
// 根据协议类型创建新的适配器实例
func (f *SSPAdapterFactory) createAdapter(sspID, protocol string) adxcore.ISSPAdapter {
	switch protocol {
	case "kuaishou":
		// Import and use Kuaishou adapter from kuaishou package
		// 从kuaishou包导入并使用快手适配器
		return kuaishou.NewKuaishouAdapter(sspID)
	case "tencent":
		// TODO: Import and use Tencent adapter when implemented
		// TODO: 实现后从tencent包导入并使用腾讯适配器
		return NewStubSSPAdapter(sspID)
	case "baidu":
		// TODO: Import and use Baidu adapter when implemented
		// TODO: 实现后从baidu包导入并使用百度适配器
		return NewStubSSPAdapter(sspID)
	case "openrtb":
		// TODO: Import and use OpenRTB adapter when implemented
		// TODO: 实现后从openrtb包导入并使用OpenRTB适配器
		return NewStubSSPAdapter(sspID)
	default:
		// Fallback to stub adapter for unknown protocols
		// 对于未知协议，回退到存根适配器
		return NewStubSSPAdapter(sspID)
	}
}

// NewStubSSPAdapter creates a new stub SSP adapter
// 创建新的存根SSP适配器
func NewStubSSPAdapter(sspID string) *StubSSPAdapter {
	return &StubSSPAdapter{sspID: sspID}
}

type StubSSPAdapter struct {
	sspID string
}

// ToInternalBidRequest converts SSP-specific bid request to internal format
// 将SSP特定的竞价请求转换为内部格式
func (a *StubSSPAdapter) ToInternalBidRequest(ctx *adxcore.BidRequestCtx, data []byte) error {
	// For stub adapter, just parse as standard OpenRTB
	// 对于存根适配器，直接解析为标准OpenRTB格式
	var bidReq admux_rtb.BidRequest
	if err := json.Unmarshal(data, &bidReq); err != nil {
		return fmt.Errorf("failed to parse bid request: %v", err)
	}

	ctx.Request = &bidReq
	return nil
}

// PackSSPResponse converts internal bid response to SSP-specific format
// 将内部竞价响应转换为SSP特定格式
func (a *StubSSPAdapter) PackSSPResponse(ctx *adxcore.BidRequestCtx) ([]byte, error) {
	if ctx.Response == nil {
		return nil, fmt.Errorf("no response to pack")
	}

	// For stub adapter, just marshal as JSON
	// 对于存根适配器，直接序列化为JSON
	return json.Marshal(ctx.Response)
}
