# ADX引擎广播逻辑技术方案文档

## 1. 概述

本文档详细描述了ADX引擎中广播请求到各DSP bidder的技术方案设计。广播逻辑是ADX引擎的核心组件，负责并发地向多个DSP发送竞价请求，并收集响应进行后续处理。

### 1.1 设计目标
- **高性能**: 支持高并发、低延迟的DSP通信
- **高可用**: 具备容错、重试和健康检查机制
- **可扩展**: 支持动态添加/移除DSP bidder
- **内存高效**: 优化内存使用，避免内存泄漏
- **可观测**: 提供完整的监控和日志记录

## 2. 架构设计

### 2.1 核心组件

```
┌─────────────────────────────────────────────────────────────┐
│                      BroadcastManager                       │
├─────────────────────────────────────────────────────────────┤
│  - 管理DSP bidder健康状态                                   │
│  - 并发广播控制                                             │
│  - 超时和重试管理                                           │
│  - 响应收集和聚合                                           │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      BidderFactory                          │
├─────────────────────────────────────────────────────────────┤
│  - 注册和管理DSP bidder实例                                 │
│  - 提供bidder发现机制                                       │
│  - 线程安全的bidder访问                                     │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                        DSP Bidder                           │
├─────────────────────────────────────────────────────────────┤
│  - 实现Bidder接口                                           │
│  - 处理HTTP请求/响应                                        │
│  - 超时和错误处理                                           │
│  - 健康状态管理                                             │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 核心接口定义

```go
// Bidder接口 - 所有DSP bidder必须实现
type Bidder interface {
    GetInfo() *BidderInfo                    // 获取bidder信息
    SendBidRequest(ctx *BidRequestCtx) ([]*BidCandidate, error) // 发送竞价请求
}

// Broadcaster接口 - 广播管理器接口
type Broadcaster interface {
    Broadcast(ctx *BidRequestCtx) ([]*BidCandidate, error)
}
```

## 3. 核心逻辑设计

### 3.1 广播流程伪代码

```
算法: BroadcastToBidders
输入: bidRequest - 竞价请求上下文
输出: []BidResponse - 所有bidder的响应集合

BEGIN
    // 1. 获取健康且可用的bidder列表
    healthyBidders ← GetHealthyBidders()

    // 2. 创建响应收集通道
    responseChan ← make(chan BidResponse, len(healthyBidders))

    // 3. 为每个bidder创建独立的上下文和超时控制
    FOR EACH bidder IN healthyBidders DO
        // 创建带超时的子上下文
        bidderCtx ← context.WithTimeout(bidRequest.Context, bidder.Timeout)

        // 4. 并发发送请求
        GO func(bidder, bidderCtx) {
            response ← BidResponse{BidderID: bidder.ID}

            START_TIMER
            candidates, err ← bidder.SendBidRequest(bidderCtx)
            response.Latency ← STOP_TIMER

            IF err != nil THEN
                response.Error ← err
                UPDATE_BIDDER_HEALTH(bidder, false)  // 标记为不健康
            ELSE
                response.Candidates ← candidates
            END IF

            responseChan ← response
        }(bidder, bidderCtx)
    END FOR

    // 5. 收集所有响应
    responses ← []BidResponse{}
    FOR i ← 0 TO len(healthyBidders) DO
        response ← <-responseChan
        responses ← append(responses, response)
    END FOR

    // 6. 返回响应集合
    RETURN responses, nil
END
```

### 3.2 超时控制策略

#### 3.2.1 分层超时设计

```
┌─────────────────────────────────────────────────────────────┐
│                    SSP级别超时 (3000ms)                     │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                广播阶段超时 (2500ms)                    │ │
│  ├─────────────────────────────────────────────────────────┤ │
│  │  ┌─────────────────────────────────────────────────────┐ │ │
│  │  │             单个DSP超时 (2000ms)                    │ │ │
│  │  └─────────────────────────────────────────────────────┘ │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

#### 3.2.2 超时配置

```yaml
# SSP配置
ssps:
  - id: "ssp-001"
    name: "示例SSP"
    timeout: 3000ms  # SSP级别总超时

# DSP bidder配置
bidders:
  - id: "dsp-001"
    name: "示例DSP"
    timeout: 2000ms  # 单个bidder超时
    retry_count: 2   # 重试次数
    retry_delay: 100ms  # 重试延迟
```

### 3.3 内存优化策略

#### 3.3.1 对象池化
- **BidRequestCtx池**: 复用请求上下文对象
- **BidCandidate池**: 复用竞价候选对象
- **HTTP连接池**: 复用HTTP客户端连接

#### 3.3.2 内存使用控制
```
算法: 内存使用监控和控制
BEGIN
    // 监控内存使用
    memoryUsage ← GetCurrentMemoryUsage()

    IF memoryUsage > WARNING_THRESHOLD THEN
        // 触发内存优化策略
        TRIGGER_MEMORY_OPTIMIZATION()
    END IF

    IF memoryUsage > CRITICAL_THRESHOLD THEN
        // 拒绝新请求，保护系统
        REJECT_NEW_REQUESTS()
    END IF
END
```

## 4. 并发控制设计

并发超时控制与retry等模块，按照功能实现到公共的pkg中，避免与业务代码耦合,利用范型等技术，避免类型转化开销
同时将并发控制、收集结果的逻辑应该可以实现成公共包；

### 4.1 Goroutine管理

```go
// 并发控制结构
type ConcurrencyController struct {
    semaphore chan struct{}  // 信号量控制并发数
    wg        sync.WaitGroup // 等待组
    mu        sync.RWMutex   // 读写锁
}

// 伪代码: 并发广播控制
算法: ConcurrentBroadcast
输入: bidders - bidder列表, request - 请求
输出: responses - 响应集合

BEGIN
    controller ← NewConcurrencyController(MAX_CONCURRENCY)
    responses ← make([]BidResponse, 0, len(bidders))

    FOR EACH bidder IN bidders DO
        controller.wg.Add(1)

        GO func(bidder) {
            // 获取信号量许可
            controller.semaphore ← struct{}{}

            // 执行bidder请求
            response ← executeBidderRequest(bidder, request)

            // 线程安全地添加响应
            controller.mu.Lock()
            responses ← append(responses, response)
            controller.mu.Unlock()

            // 释放信号量
            ←controller.semaphore
            controller.wg.Done()
        }(bidder)
    END FOR

    // 等待所有goroutine完成
    controller.wg.Wait()
    RETURN responses
END
```

### 4.2 QPS限制

```go
// QPS限制器
type QPSLimiter struct {
    rateLimiter *rate.Limiter
    lastReset   time.Time
    mu          sync.Mutex
}

// 伪代码: QPS控制
算法: CheckAndUpdateQPS
输入: bidder - DSP bidder
输出: bool - 是否允许请求

BEGIN
    limiter ← GetBidderLimiter(bidder.ID)

    // 检查QPS限制
    IF NOT limiter.Allow() THEN
        LOG_WARNING("QPS limit exceeded for bidder", bidder.ID)
        RETURN false
    END IF

    RETURN true
END
```

## 5. 错误处理和重试机制

### 5.1 错误分类

```go
// 错误类型定义
type BroadcastError struct {
    Type     ErrorType  // 错误类型
    BidderID string     // bidder标识
    Message  string     // 错误信息
    Retryable bool      // 是否可重试
}

type ErrorType int

const (
    TimeoutError ErrorType = iota    // 超时错误
    NetworkError                     // 网络错误
    ProtocolError                    // 协议错误
    RateLimitError                   // 限流错误
    InternalError                    // 内部错误
)
```

### 5.2 重试策略

```
算法: RetryWithBackoff
输入: operation - 操作函数, maxRetries - 最大重试次数
输出: result - 操作结果, error - 错误

BEGIN
    lastErr ← nil

    FOR retry ← 0 TO maxRetries DO
        result, err ← operation()

        IF err == nil THEN
            RETURN result, nil
        END IF

        // 检查错误是否可重试
        IF NOT IsRetryableError(err) THEN
            RETURN nil, err
        END IF

        lastErr ← err

        // 指数退避延迟
        delay ← CalculateBackoffDelay(retry)
        SLEEP(delay)
    END FOR

    RETURN nil, lastErr
END
```

## 6. 健康检查和熔断机制

### 6.1 健康检查策略

```go
// 健康检查器
type HealthChecker struct {
    checkInterval time.Duration
    failureThreshold int
    successThreshold int
    mu sync.RWMutex
    status map[string]*HealthStatus
}

type HealthStatus struct {
    Healthy       bool
    FailureCount  int
    SuccessCount  int
    LastCheck     time.Time
}

// 伪代码: 健康状态更新
算法: UpdateHealthStatus
输入: bidderID - bidder标识, success - 是否成功

BEGIN
    status ← GetHealthStatus(bidderID)

    IF success THEN
        status.SuccessCount ← status.SuccessCount + 1
        status.FailureCount ← 0

        IF status.SuccessCount >= SUCCESS_THRESHOLD THEN
            status.Healthy ← true
        END IF
    ELSE
        status.FailureCount ← status.FailureCount + 1
        status.SuccessCount ← 0

        IF status.FailureCount >= FAILURE_THRESHOLD THEN
            status.Healthy ← false
        END IF
    END IF

    status.LastCheck ← time.Now()
END
```

### 6.2 熔断器模式

```
状态机: 熔断器状态转换

CLOSED (正常) → (失败次数 > 阈值) → OPEN (熔断)
    ↑                                      |
    |                                  (超时后)
    |                                      ↓
    ←────────── HALF-OPEN (试探) ←──────────
            (成功次数 > 阈值)
```

## 7. 性能优化策略

### 7.1 连接池优化

```go
// HTTP连接池配置
type HTTPPoolConfig struct {
    MaxIdleConns        int           // 最大空闲连接数
    MaxIdleConnsPerHost int           // 每主机最大空闲连接数
    IdleConnTimeout     time.Duration // 空闲连接超时
    DisableCompression  bool          // 禁用压缩
}

// 伪代码: 连接池初始化
算法: InitializeHTTPPool
输入: config - 连接池配置
输出: *http.Client - HTTP客户端

BEGIN
    transport ← &http.Transport{
        MaxIdleConns:        config.MaxIdleConns,
        MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
        IdleConnTimeout:     config.IdleConnTimeout,
        DisableCompression:  config.DisableCompression,
    }

    RETURN &http.Client{
        Transport: transport,
        Timeout:   DEFAULT_TIMEOUT,
    }
END
```

### 7.2 响应处理优化

```go
// 批量响应处理
type BatchProcessor struct {
    batchSize    int
    processFunc  func([]BidResponse)
    buffer       []BidResponse
    mu           sync.Mutex
}

// 伪代码: 批量处理响应
算法: ProcessResponsesInBatch
输入: responses - 响应集合

BEGIN
    // 分批处理响应，减少锁竞争
    FOR i ← 0; i < len(responses); i += BATCH_SIZE DO
        end ← min(i + BATCH_SIZE, len(responses))
        batch ← responses[i:end]

        GO processBatch(batch)
    END FOR
END
```

## 8. 监控和可观测性

### 8.1 关键指标

```go
// 监控指标定义
type BroadcastMetrics struct {
    TotalRequests    prometheus.Counter   // 总请求数
    SuccessResponses prometheus.Counter   // 成功响应数
    FailedResponses  prometheus.Counter   // 失败响应数
    ResponseLatency  prometheus.Histogram // 响应延迟
    ActiveBidders    prometheus.Gauge     // 活跃bidder数
    QPSLimitHits     prometheus.Counter   // QPS限制命中数
}
```

### 8.2 日志策略

```go
// 结构化日志记录
func (bm *BroadcastManager) logBroadcastResult(
    ctx *BidRequestCtx,
    responses []BidResponse,
    duration time.Duration,
) {
    logger.Info().
        Str("request_id", GetRequestID(ctx)).
        Int("total_bidders", len(responses)).
        Int("success_count", countSuccess(responses)).
        Int("failure_count", countFailures(responses)).
        Dur("total_duration", duration).
        Msg("broadcast_completed")
}
```

## 9. 配置管理

### 9.1 动态配置

```yaml
broadcast:
  max_concurrency: 50           # 最大并发数
  default_timeout: "2000ms"     # 默认超时时间
  health_check_interval: "30s"  # 健康检查间隔

  circuit_breaker:
    failure_threshold: 5        # 失败阈值
    success_threshold: 3        # 成功阈值
    timeout: "60s"              # 熔断超时

  retry_policy:
    max_retries: 2              # 最大重试次数
    backoff_multiplier: 2.0     # 退避乘数
    initial_delay: "100ms"      # 初始延迟
    max_delay: "1s"             # 最大延迟
```

## 10. 总结

本技术方案提供了ADX引擎广播逻辑的完整设计，包括：

1. **高性能并发架构**: 支持高并发DSP通信
2. **精细化超时控制**: 分层超时策略确保系统稳定性
3. **完善的错误处理**: 重试、熔断和健康检查机制
4. **内存优化**: 对象池化和内存使用控制
5. **全面监控**: 完整的指标和日志记录
6. **安全可靠**: 请求验证和资源保护

该设计遵循广告竞价的最佳实践，确保系统在高负载下仍能保持稳定和高效运行。
