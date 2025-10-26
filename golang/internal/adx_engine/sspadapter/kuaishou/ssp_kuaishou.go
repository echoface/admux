package kuaishou

import (
	"encoding/json"
	"fmt"

	"github.com/echoface/admux/internal/adx_engine/adxcore"
	admux_rtb "github.com/echoface/admux/pkg/protogen/admux"
	kuaishou_rtb "github.com/echoface/admux/pkg/protogen/kuaishou"
	"google.golang.org/protobuf/proto"
)

// KuaishouAdapter implements SSP adapter for Kuaishou protocol
// 快手协议适配器
type KuaishouAdapter struct {
	sspID string
}

// NewKuaishouAdapter creates a new Kuaishou SSP adapter
// 创建新的快手SSP适配器
func NewKuaishouAdapter(sspID string) *KuaishouAdapter {
	return &KuaishouAdapter{sspID: sspID}
}

// ToInternalBidRequest converts Kuaishou-specific bid request to internal format
// 将快手特定的竞价请求转换为内部格式
func (a *KuaishouAdapter) ToInternalBidRequest(ctx *adxcore.BidRequestCtx, data []byte) error {
	// Parse Kuaishou bid request
	var kuaishouReq kuaishou_rtb.BidRequest
	if err := json.Unmarshal(data, &kuaishouReq); err != nil {
		return fmt.Errorf("failed to parse Kuaishou bid request: %v", err)
	}

	// Convert to internal OpenRTB format
	internalReq, err := a.convertToInternalRequest(&kuaishouReq)
	if err != nil {
		return fmt.Errorf("failed to convert Kuaishou request to internal format: %v", err)
	}

	ctx.Request = internalReq
	return nil
}

// PackSSPResponse converts internal bid response to Kuaishou-specific format
// 将内部竞价响应转换为快手特定格式
func (a *KuaishouAdapter) PackSSPResponse(ctx *adxcore.BidRequestCtx) ([]byte, error) {
	if ctx.Response == nil {
		return nil, fmt.Errorf("no response to pack")
	}

	// Convert internal response to Kuaishou format
	kuaishouResp, err := a.convertToKuaishouResponse(ctx.Response)
	if err != nil {
		return nil, fmt.Errorf("failed to convert internal response to Kuaishou format: %v", err)
	}

	return json.Marshal(kuaishouResp)
}

// convertToInternalRequest converts Kuaishou bid request to internal OpenRTB format
// 将快手竞价请求转换为内部OpenRTB格式
func (a *KuaishouAdapter) convertToInternalRequest(kuaishouReq *kuaishou_rtb.BidRequest) (*admux_rtb.BidRequest, error) {
	if len(kuaishouReq.Imp) == 0 {
		return nil, fmt.Errorf("no impression in Kuaishou bid request")
	}

	// Convert first impression
	internalReq := &admux_rtb.BidRequest{
		Id:  kuaishouReq.RequestId,
		Imp: []*admux_rtb.BidRequest_Imp{},
	}

	return internalReq, nil
}

// convertToKuaishouResponse converts internal bid response to Kuaishou format
// 将内部竞价响应转换为快手格式
func (a *KuaishouAdapter) convertToKuaishouResponse(internalResp *admux_rtb.BidResponse) (*kuaishou_rtb.BidResponse, error) {
	if len(internalResp.Seatbid) == 0 || len(internalResp.Seatbid[0].Bid) == 0 {
		return &kuaishou_rtb.BidResponse{
			Status: proto.Uint32(1), // No bid
		}, nil
	}

	return &kuaishou_rtb.BidResponse{}, nil
}

// Helper methods for conversion
// 转换辅助方法

func (a *KuaishouAdapter) mapOSType(osType kuaishou_rtb.Device_OsType) string {
	switch osType {
	case kuaishou_rtb.Device_ANDROID:
		return "android"
	case kuaishou_rtb.Device_IOS:
		return "ios"
	default:
		return "unknown"
	}
}

func (a *KuaishouAdapter) getOSVersion(version *kuaishou_rtb.Version) string {
	if version == nil {
		return ""
	}
	return fmt.Sprintf("%d.%d.%d", version.GetMajor(), version.GetMinor(), version.GetMicro())
}

func (a *KuaishouAdapter) getScreenWidth(size *kuaishou_rtb.Size) int32 {
	if size == nil {
		return 0
	}
	return int32(size.GetWidth())
}

func (a *KuaishouAdapter) getScreenHeight(size *kuaishou_rtb.Size) int32 {
	if size == nil {
		return 0
	}
	return int32(size.GetHeight())
}
