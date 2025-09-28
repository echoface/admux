package adxserver

import (
	"context"
	"io"

	"github.com/echoface/admux/api/gen/admux/openrtb"
	"github.com/echoface/admux/internal/adxcore"
	"github.com/echoface/admux/internal/sspadapter"
	"github.com/echoface/admux/pkg/utils"
)

type PipelineStage interface {
	Process(ctx adxcore.BidRequestCtx) error
}

type AdxServer struct {
	stages []PipelineStage
	appCtx *AdxServerContext
}

func NewAdxServer(appCtx *AdxServerContext) *AdxServer {
	return &AdxServer{
		stages: []PipelineStage{
			&SSIDValidationStage{},
			&BidProcessingStage{appCtx: appCtx},
		},
		appCtx: appCtx,
	}
}

func (s *AdxServer) ProcessBid(ctx context.Context, bidReq *openrtb.BidRequest) (any, error) {
	// TODO:(gonghuan) 补全系统内相关基础信息； 将一些配置和用户dmp数据准备好

	for _, stage := range s.stages {
		err := stage.Process(ctx)
		if err != nil {
			return nil, err
		}
		currentData = result
	}

	return currentData, nil
}

// TODO: 校验BidRequest 中的请求是否满足要求; 将这个stage独立文件实现
type SSIDValidationStage struct{}

func (s *SSIDValidationStage) Process(ctx adxcore.BidRequestCtx) error {
	ssid, ok := ctx.Value("ssid").(string)
	if !ok || ssid == "" {
		return ErrMissingSSID
	}

	return nil
}

// TODO: 校验BidRequest 中的请求是否满足要求; 将这个stage独立文件实现
type BidProcessingStage struct {
	appCtx *AdxServerContext
}

func (s *BidProcessingStage) Process(ctx *adxcore.BidRequestCtx) (any, error) {
	// 创建广播管理器
	broadcastManager := NewBroadcastManager(s.appCtx)

	// 向所有DSP bidder广播竞价请求
	responses, err := broadcastManager.BroadcastToBidders(ctx)
	if err != nil {
		return nil, err
	}
	utils.Ignore(responses)
	utils.IgnoreErr(err, "broadcast error")

	// 选择获胜的出价

	// 构建最终响应
	return map[string]any{}, nil
}

var ErrMissingSSID = &AdxError{Message: "missing ssid parameter", Code: "MISSING_SSID"}

type AdxError struct {
	Message string
	Code    string
}

func (e *AdxError) Error() string {
	return e.Message
}
