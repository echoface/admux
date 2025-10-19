# admux adx_engine

一个广告adx竞价平台，使用openRTB标准协议

通过对接层ssp-adapter 接入市面上主流的ssp，比如小米、快手、巨量引擎、猫眼等；接入后转为标准的openrtb协议进入bidding流程
通过获取相关特征信息（人群包、mapping等等）之后进行竞价方的召回(recall bidder)，获取到满足条件的bidder之后并发对每一个bidder
请求，请求过程中完成openrtb 到 bidder 协议的转化 并发起请求，拿到响应后转化成openrtb的response 和相关信息组成上下为作为BidCandidate;
若干candiates进行比价(粗排/混排/精排)得到与请求要求的bid数量将这N个结果进入ssp 的response 打包环节； 生成ssp 所需要的response格式

## 流量来源SSP adapter
./internal/ssp

    - 将ssp的竞价数据转化生成BidReqCtx

## adx 服务实现

- 主程逻辑链路
./internal/adx, 依赖adxcore


- adxcore 核心bid上下文
./internal/adxcore； 定义了主程序流程中核心的一些概念;

    - BidReqCtx: 竞价上下文
        - ssp
        - BidRequest
            - impls
        - ServerCtx
            - redis缓存呢你
            - 人群定向能力
            - cookie mapping
            - bidder数据管理
    - Candidate
        - BidResponse

## dsp 服务实现 adapter
./internal/bidder

    - BidderCtx：竞价方上下文
        - BidderInfo 竞价方信息,地址/qps 等等
        - BidRequest
        - BidResponse


## Packer 竞价响应打包
    -
