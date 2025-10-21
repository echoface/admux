# admux adx_engine

一个广告adx竞价平台，使用openRTB标准协议

通过对接层ssp-adapter 接入市面上主流的ssp，比如小米、快手、巨量引擎、猫眼等；接入后转为标准的openrtb协议进入bidding流程
通过获取相关特征信息（人群包、mapping等等）之后进行竞价方的召回(recall bidder)，获取到满足条件的bidder之后并发对每一个bidder
请求，请求过程中完成openrtb 到 bidder 协议的转化 并发起请求，拿到响应后转化成openrtb的response 和相关信息组成上下为作为BidCandidate;
若干candiates进行比价(粗排/混排/精排)得到与请求要求的bid数量将这N个结果进入ssp 的response 打包环节； 生成ssp 所需要的response格式


## 代码架构

### 依赖关系
本项目采用分层架构； 虽然在分层架构、IOC倒置、垂直切分等方案， 但是考虑到研发背景与代码控制能力综合考虑采用分层架构；实现代码过程中利用优势，尽量避免这种架构的缺点，取长补短；
``` 结构示例
adxserver → sspadapter → adxcore
    ↓           ↓           ↓
 (上层)      (中间层)     (底层)
 ```
✅ 清晰的架构层次：符合依赖倒置原则
✅ 核心业务独立：adxcore 不依赖外部，可独立测试
✅ 易于扩展：新增 SSP 协议不影响核心逻辑
✅ 技术栈隔离：协议转换与业务逻辑分离

缺点
❌ 间接调用：adxserver 需要协调两个依赖
❌ 数据转换开销：需要在层间传递数据
❌ 接口膨胀：可能需要定义较多的接口

#### 推荐方案对比
维度	      方案一（分层）	方案二（IoC）	方案三（垂直）	方案四（事件）
架构清晰度	  ⭐⭐⭐⭐⭐	⭐⭐⭐⭐	  ⭐⭐⭐	      ⭐⭐
开发效率	  ⭐⭐⭐⭐	  ⭐⭐⭐⭐	  ⭐⭐⭐⭐	   ⭐⭐
测试便利性	 ⭐⭐⭐⭐	   ⭐⭐⭐⭐⭐	⭐⭐⭐	      ⭐⭐
性能	     ⭐⭐⭐⭐	   ⭐⭐⭐⭐	   ⭐⭐⭐⭐⭐	 ⭐⭐⭐
扩展性	    ⭐⭐⭐⭐	  ⭐⭐⭐⭐⭐	 ⭐⭐⭐	     ⭐⭐⭐⭐⭐
团队协作	 ⭐⭐⭐⭐	   ⭐⭐⭐⭐	    ⭐⭐⭐⭐⭐	 ⭐⭐
```

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
