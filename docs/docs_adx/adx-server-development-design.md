# ADX服务器开发设计文档

## 1. 项目概述

### 1.1 项目背景
ADMUX是一个面向中国国内竞价广告市场的Ad Exchange (ADX)平台，采用OpenRTB标准协议，实现高吞吐、低延迟的实时竞价（RTB）系统。平台通过SSP-adapter接入主流流量供应方（小米、快手、巨量引擎、猫眼等），通过DSP-adapter对接广告需求方，完成完整的竞价交易流程。

### 1.2 核心目标
- 实现QPS ≥ 10000，端到端延迟 ≤ 50ms的高性能RTB系统
- 支持主流SSP平台的协议适配和标准化转换
- 提供稳定的流量分发和用户定向能力
- 最大化广告请求填充率和变现效率（eCPM）
- 支持修正的第一价格竞价机制

### 1.3 技术架构概览
```
SSP平台 → SSP Adapter → ADX Core → DSP Adapter → DSP平台
              ↓            ↑            ↓
         协议转换    竞价决策引擎    协议转换
              ↓            ↑            ↓
         用户画像      流量定向      响应处理
```

## 2. 技术选型

### 2.1 核心技术栈
- **开发语言**: Go 1.24.5
  - 高并发性能，低延迟
  - 丰富的标准库和生态
  - 简洁的语法和优秀的工具链

- **Web框架**: Gin v1.11.0
  - 高性能HTTP框架
  - 轻量级，路由性能优秀
  - 支持中间件机制

- **序列化协议**: Google Protobuf v1.36.9
  - 高效的二进制序列化
  - 强类型接口定义
  - 跨语言支持

- **缓存系统**: Redis
  - 高性能内存数据库
  - 支持多种数据结构
  - 丰富的持久化选项

### 2.2 依赖库选择
- **监控指标**: Prometheus client_golang v1.23.2
- **HTTP客户端**: 自研高性能客户端（支持连接池、超时控制）
- **配置管理**: Viper（动态配置加载）
- **日志系统**: Zap（高性能结构化日志）

### 2.3 部署架构
- **容器化**: Docker + Docker Compose
- **负载均衡**: Nginx/HAProxy
- **配置管理**: 环境变量 + 配置文件
- **消息队列**: Kafka（竞价日志、审计数据）

## 3. 模块设计与开发计划

### 3.1 开发顺序规划
按照依赖关系和优先级，按以下顺序进行开发：

1. **核心数据结构定义** (Week 1-2)
2. **配置管理系统** (Week 2-3)
3. **SSP适配器层** (Week 3-5)
4. **ADX核心引擎** (Week 5-8)
5. **DSP适配器层** (Week 8-10)
6. **竞价决策引擎** (Week 10-12)
7. **监控与日志系统** (Week 12-13)
8. **性能优化与测试** (Week 13-14)

### 3.2 核心模块详细设计

#### 3.2.1 核心数据结构 (internal/adxcore)

**开发目标**: 定义系统中所有核心数据结构和接口

**关键数据结构**:
```go
// 竞价请求上下文
type BidReqCtx struct {
    RequestID     string                 // 请求唯一标识
    SSP           SSP                    // 流量来源SSP信息
    BidRequest    openrtb.BidRequest     // 标准化竞价请求
    ServerCtx     *ServerContext         // 服务上下文
    Timestamp     time.Time             // 请求时间戳
}

// 服务上下文
type ServerContext struct {
    RedisClient   redis.Client          // Redis客户端
    AudienceMgr   AudienceManager      // 人群定向管理
    CookieMapping CookieMappingService // Cookie映射服务
    BidderMgr     BidderManager        // 竞价方管理
}

// 竞价候选
type Candidate struct {
    BidResponse   openrtb.BidResponse  // 竞价响应
    Bidder        Bidder               // 竞价方信息
    ECPM          float64              // 有效千次曝光收入
    Score         float64              // 综合评分
}

// SSP信息
type SSP struct {
    ID          string    // SSP唯一标识
    Name        string    // SSP名称
    Protocol    string    // 支持协议版本
    Endpoint    string    // 接入端点
    QPSLimit    int       // QPS限制
    Timeout     time.Duration // 超时设置
}

// Bidder信息
type Bidder struct {
    ID          string    // Bidder唯一标识
    Name        string    // Bidder名称
    Endpoint    string    // 服务端点
    QPSLimit    int       // QPS限制
    Timeout     time.Duration // 超时设置
    AuthToken   string    // 认证令牌
}
```

**开发内容**:
1. 定义完整的OpenRTB协议数据结构
2. 实现BidReqCtx的生命周期管理
3. 设计ServerCtx的初始化和资源管理
4. 实现Candidate的排序和筛选逻辑
5. 定义SSP和Bidder的配置结构

#### 3.2.2 配置管理系统 (internal/config)

**开发目标**: 实现动态配置加载和管理

**配置结构**:
```yaml
server:
  port: 8080
  read_timeout: 100ms
  write_timeout: 100ms
  max_connections: 10000

redis:
  addr: "localhost:6379"
  pool_size: 100
  min_idle_conns: 10

ssps:
  - id: "xiaomi"
    name: "小米"
    endpoint: "/bid/xiaomi"
    protocol: "openrtb"
    qps_limit: 1000

  - id: "kuaishou"
    name: "快手"
    endpoint: "/bid/kuaishou"
    protocol: "custom"
    qps_limit: 2000

bidders:
  - id: "dsp1"
    name: "DSP服务商1"
    endpoint: "http://dsp1.example.com/bid"
    timeout: 50ms
    qps_limit: 500

routing:
  default_floor_price: 0.5
  profit_margin: 0.15
  min_bid_increment: 0.01
```

**开发内容**:
1. 实现Viper配置加载器
2. 支持环境变量覆盖
3. 实现配置热更新机制
4. 配置验证和默认值处理
5. 配置变更通知机制

#### 3.2.3 SSP适配器层 (internal/sspadapter)

**开发目标**: 实现多SSP平台的协议适配和标准化

**架构设计**:
```go
// SSP适配器接口
type SSPAdapter interface {
    ParseRequest(req *http.Request) (*BidReqCtx, error)
    BuildResponse(ctx *BidReqCtx, candidate *Candidate) (*http.Response, error)
    ValidateRequest(req *http.Request) error
}

// 适配器工厂
type AdapterFactory interface {
    CreateAdapter(sspID string) (SSPAdapter, error)
    RegisterAdapter(sspID string, adapter SSPAdapter)
}

// 基础适配器实现
type BaseAdapter struct {
    config   *SSPConfig
    logger   logger.Logger
    metrics  metrics.Recorder
}
```

**各SSP适配器开发计划**:

1. **OpenRTB标准适配器** (优先级: 高)
   - 支持OpenRTB 2.5/2.6协议
   - 标准字段映射和验证
   - 错误处理和日志记录

2. **小米适配器** (优先级: 高)
   - 解析小米私有协议格式
   - 设备ID映射（OAID/IDFA）
   - 特殊字段处理

3. **快手适配器** (优先级: 高)
   - 快手协议解析
   - 用户标签处理
   - 版位信息映射

4. **巨量引擎适配器** (优先级: 中)
   - 巨量协议适配
   - 创意格式转换

5. **猫眼适配器** (优先级: 低)
   - 猫眼协议支持
   - 特殊业务逻辑处理

**开发内容**:
1. 实现适配器注册和发现机制
2. 协议解析和转换逻辑
3. 请求验证和错误处理
4. 性能监控和指标收集
5. 单元测试和集成测试

#### 3.2.4 ADX核心引擎 (internal/adx)

**开发目标**: 实现竞价请求处理的核心流程

**核心流程设计**:
```go
// ADX服务主处理器
type ADXService struct {
    config       *Config
    sspAdapter   SSPAdapterManager
    dmpService   DMPService
    router       TrafficRouter
    broadcaster  BidBroadcaster
    selector     CandidateSelector
    packer       ResponsePacker
}

// 主要处理流程
func (s *ADXService) HandleBidRequest(w http.ResponseWriter, r *http.Request) {
    // 1. 协议解析和验证
    ctx, err := s.sspAdapter.ParseRequest(r)

    // 2. 用户特征增强
    err = s.dmpService.EnrichUserFeatures(ctx)

    // 3. 流量分发定向
    targetBidders := s.RouteToBidders(ctx)

    // 4. 并发竞价请求
    candidates := s.broadcaster.BroadcastBid(ctx, targetBidders)

    // 5. 竞价决策
    winner := s.selector.SelectWinner(candidates, ctx)

    // 6. 响应打包
    response := s.packer.BuildResponse(ctx, winner)

    // 7. 返回响应
    s.sendResponse(w, response)
}
```

**开发内容**:
1. HTTP请求路由和中间件
2. 竞价流程编排和错误处理
3. 并发控制和超时管理
4. 性能监控和指标收集
5. 链路追踪和日志记录

#### 3.2.5 DSP适配器层 (internal/bidder)

**开发目标**: 实现与DSP Bidder的通信和协议适配

**架构设计**:
```go
// Bidder适配器接口
type BidderAdapter interface {
    BuildRequest(ctx *BidReqCtx) (*http.Request, error)
    ParseResponse(resp *http.Response) (*Candidate, error)
    GetEndpoint() string
    GetTimeout() time.Duration
}

// Bidder管理器
type BidderManager interface {
    GetAvailableBidders(ctx *BidReqCtx) []Bidder
    GetBidderAdapter(bidderID string) (BidderAdapter, error)
    CheckBidderHealth(bidderID string) bool
}

// 并发竞价广播器
type BidBroadcaster interface {
    BroadcastBid(ctx *BidReqCtx, bidders []Bidder) []*Candidate
    BroadcastBidAsync(ctx *BidReqCtx, bidders []Bidder) <-chan *Candidate
}
```

**开发内容**:
1. OpenRTB协议请求构建
2. 并发HTTP客户端实现
3. 响应解析和错误处理
4. 连接池和超时控制
5. 健康检查和熔断机制
6. 性能监控和指标收集

#### 3.2.6 竞价决策引擎 (internal/adxcore/decision)

**开发目标**: 实现智能的竞价决策和排序算法

**决策流程**:
```go
// 竞价决策器
type BidDecisionEngine struct {
    scorer       CandidateScorer
    sorter       CandidateSorter
    validator    CandidateValidator
    pricer       PricingEngine
}

// 竞价决策流程
func (e *BidDecisionEngine) MakeDecision(
    candidates []*Candidate,
    ctx *BidReqCtx,
) *DecisionResult {
    // 1. 候选过滤
    validCandidates := e.validator.Filter(candidates, ctx)

    // 2. 评分计算
    for _, candidate := range validCandidates {
        candidate.Score = e.scorer.CalculateScore(candidate, ctx)
        candidate.ECPM = e.calculateECPM(candidate, ctx)
    }

    // 3. 排序
    sortedCandidates := e.sorter.Sort(validCandidates)

    // 4. 胜出判定
    winner := e.selectWinner(sortedCandidates, ctx)

    // 5. 定价计算
    finalPrice := e.pricer.CalculateFinalPrice(winner, sortedCandidates, ctx)

    return &DecisionResult{
        Winner:    winner,
        FinalPrice: finalPrice,
    }
}
```

**评分算法设计**:
```go
// 综合评分计算
func (s *CompositeScorer) CalculateScore(candidate *Candidate, ctx *BidReqCtx) float64 {
    // 基础评分 (60%): 出价金额
    baseScore := candidate.BidResponse.Bid.Price * 0.6

    // 预估CTR评分 (30%): pCTR模型预测
    ctrScore := s.pCTREngine.PredictCTR(candidate, ctx) * candidate.BidResponse.Bid.Price * 0.3

    // 广告质量评分 (10%): 创意质量、历史表现等
    qualityScore := s.qualityModel.CalculateQualityScore(candidate, ctx) * 0.1

    return baseScore + ctrScore + qualityScore
}
```

**定价引擎设计**:
```go
// 修正第一价格定价引擎
type ModifiedFirstPricePricer struct {
    profitMargin     float64  // 利润空间
    minIncrement     float64  // 最小加价
    floorPrice       float64  // 底价
}

func (p *ModifiedFirstPricePricer) CalculateFinalPrice(
    winner *Candidate,
    allCandidates []*Candidate,
    ctx *BidReqCtx,
) float64 {
    // 获取次高价
    secondHighestPrice := p.getSecondHighestPrice(winner, allCandidates)

    // 计算最终价格 = max(次高价 + 最小加价, 底价) * (1 + 利润空间)
    finalPrice := math.Max(secondHighestPrice+p.minIncrement, p.floorPrice)
    finalPrice = finalPrice * (1 + p.profitMargin)

    // 不超过获胜方出价
    finalPrice = math.Min(finalPrice, winner.BidResponse.Bid.Price)

    return finalPrice
}
```

**开发内容**:
1. 多维度评分算法实现
2. 候选过滤和排序逻辑
3. 定价引擎和利润计算
4. A/B测试框架支持
5. 算法性能优化

#### 3.2.7 监控与日志系统

**开发目标**: 实现全面的系统监控和可观测性

**监控指标设计**:
```go
// 核心业务指标
var (
    // QPS指标
    RequestTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "adx_requests_total",
            Help: "Total number of bid requests received",
        },
        []string{"ssp", "status"},
    )

    // 延迟指标
    RequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "adx_request_duration_seconds",
            Help:    "Request processing duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"ssp", "stage"},
    )

    // 竞价成功指标
    BidSuccessRate = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "adx_bid_success_rate",
            Help: "Bid success rate by SSP",
        },
        []string{"ssp"},
    )

    // 收入指标
    RevenueTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "adx_revenue_total",
            Help: "Total revenue generated",
        },
        []string{"ssp", "bidder"},
    )
)
```

**日志结构设计**:
```go
// 结构化日志
type BidLog struct {
    Timestamp    time.Time              `json:"timestamp"`
    RequestID    string                 `json:"request_id"`
    SSP          string                 `json:"ssp"`
    Bidder       string                 `json:"bidder"`
    FloorPrice   float64                `json:"floor_price"`
    WinPrice     float64                `json:"win_price"`
    ECPM         float64                `json:"ecpm"`
    Duration     time.Duration          `json:"duration"`
    Status       string                 `json:"status"`
    ErrorCode    string                 `json:"error_code,omitempty"`
}
```

**开发内容**:
1. Prometheus指标收集器
2. 结构化日志记录器
3. 分布式链路追踪
4. 告警规则和通知
5. 监控仪表板配置

## 4. 性能优化策略

### 4.1 并发处理优化
- 使用Go的Goroutine池控制并发度
- 实现工作窃取调度器
- 优化锁竞争和内存分配

### 4.2 缓存策略
- Redis集群部署，数据分片存储
- 热点数据本地缓存（LRU Cache）
- 预加载和缓存预热机制

### 4.3 数据库优化
- 连接池配置优化
- 读写分离和分库分表
- 索引优化和查询优化

### 4.4 网络优化
- HTTP/2支持和连接复用
- 响应数据压缩（Gzip/Brotli）
- CDN加速和静态资源优化

## 5. Docker容器化部署

### 5.1 Docker镜像构建

**多阶段构建Dockerfile**:
```dockerfile
# 构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要工具
RUN apk add --no-cache git ca-certificates tzdata

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o adx-server ./cmd/adx_server

# 运行阶段
FROM alpine:latest

# 安装ca证书和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1001 -S adx && \
    adduser -u 1001 -S adx -G adx

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/adx-server .

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 设置文件权限
RUN chown -R adx:adx /app

# 切换到非root用户
USER adx

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动应用
CMD ["./adx-server"]
```

### 5.2 Docker Compose配置

**docker-compose.yml**:
```yaml
version: '3.8'

services:
  # Redis缓存服务
  redis:
    image: redis:7-alpine
    container_name: admux-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
      - ./configs/redis.conf:/usr/local/etc/redis/redis.conf
    command: redis-server /usr/local/etc/redis/redis.conf
    networks:
      - admux-network

  # ADX主服务
  adx-server:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: admux-adx-server
    ports:
      - "8080:8080"
    environment:
      - REDIS_URL=redis://redis:6379
      - LOG_LEVEL=info
      - SERVER_PORT=8080
    volumes:
      - ./configs:/app/configs
      - ./logs:/app/logs
    depends_on:
      - redis
    networks:
      - admux-network
    restart: unless-stopped

  # 监控服务 - Prometheus
  prometheus:
    image: prom/prometheus:latest
    container_name: admux-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'
    networks:
      - admux-network

  # 监控服务 - Grafana
  grafana:
    image: grafana/grafana:latest
    container_name: admux-grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources
    depends_on:
      - prometheus
    networks:
      - admux-network

  # 日志收集服务
  filebeat:
    image: docker.elastic.co/beats/filebeat:8.5.0
    container_name: admux-filebeat
    user: root
    volumes:
      - ./logs:/app/logs
      - ./monitoring/filebeat.yml:/usr/share/filebeat/filebeat.yml:ro
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
    environment:
      - ELASTICSEARCH_HOST=elasticsearch:9200
    depends_on:
      - adx-server
    networks:
      - admux-network

  # Nginx负载均衡
  nginx:
    image: nginx:alpine
    container_name: admux-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./deployments/nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./deployments/nginx/ssl:/etc/nginx/ssl
    depends_on:
      - adx-server
    networks:
      - admux-network
    restart: unless-stopped

volumes:
  redis_data:
  prometheus_data:
  grafana_data:

networks:
  admux-network:
    driver: bridge
```

### 5.3 构建和部署脚本

**构建脚本 (build.sh)**:
```bash
#!/bin/bash

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 函数定义
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Docker和Docker Compose
check_dependencies() {
    log_info "检查依赖..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装，请先安装Docker"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose未安装，请先安装Docker Compose"
        exit 1
    fi

    log_info "依赖检查完成"
}

# 构建Docker镜像
build_images() {
    log_info "开始构建Docker镜像..."

    # 构建ADX服务镜像
    docker build -t admux/adx-server:latest .

    if [ $? -eq 0 ]; then
        log_info "ADX服务镜像构建成功"
    else
        log_error "ADX服务镜像构建失败"
        exit 1
    fi
}

# 启动服务
start_services() {
    log_info "启动服务..."

    # 创建必要的目录
    mkdir -p logs configs monitoring

    # 启动服务
    docker-compose up -d

    if [ $? -eq 0 ]; then
        log_info "服务启动成功"
    else
        log_error "服务启动失败"
        exit 1
    fi
}

# 健康检查
health_check() {
    log_info "等待服务启动完成..."
    sleep 10

    # 检查ADX服务
    if curl -f http://localhost:8080/health &> /dev/null; then
        log_info "ADX服务运行正常"
    else
        log_warn "ADX服务可能未正常启动，请检查日志"
    fi

    # 检查其他服务
    log_info "服务状态:"
    docker-compose ps
}

# 主函数
main() {
    log_info "开始构建和部署ADMUX系统..."

    check_dependencies
    build_images
    start_services
    health_check

    log_info "部署完成！"
    log_info "访问地址:"
    log_info "  ADX服务: http://localhost:8080"
    log_info "  Prometheus: http://localhost:9090"
    log_info "  Grafana: http://localhost:3000 (admin/admin)"
}

# 执行主函数
main "$@"
```

**部署脚本 (deploy.sh)**:
```bash
#!/bin/bash

set -e

# 配置变量
APP_NAME="admux"
VERSION=${1:-latest}
ENVIRONMENT=${2:-production}

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 滚动更新
rolling_update() {
    log_info "开始滚动更新..."

    # 拉取最新镜像
    docker-compose pull adx-server

    # 逐个更新实例
    docker-compose up -d --no-deps adx-server

    # 健康检查
    sleep 10
    if curl -f http://localhost:8080/health &> /dev/null; then
        log_info "滚动更新成功"
    else
        log_error "滚动更新失败，开始回滚..."
        docker-compose rollback
        exit 1
    fi
}

# 清理旧镜像
cleanup() {
    log_info "清理旧镜像..."
    docker image prune -f
    docker volume prune -f
}

# 主函数
main() {
    log_info "开始部署 $APP_NAME:$VERSION 到 $ENVIRONMENT 环境..."

    rolling_update
    cleanup

    log_info "部署完成！"
}

main "$@"
```

## 6. 质量保证

### 6.1 测试策略
- **单元测试**: 覆盖率 ≥ 80%
- **集成测试**: 端到端流程验证
- **性能测试**: 压力测试和基准测试
- **兼容性测试**: 多SSP协议兼容性

### 6.2 代码质量
- 静态代码分析（golangci-lint）
- 代码审查流程
- 持续集成和自动化部署

### 6.3 安全考虑
- API认证和授权
- 敏感数据加密存储
- SQL注入和XSS防护
- 访问日志和审计

## 7. 项目里程碑

### Phase 1: 基础架构 (Week 1-4)
- [ ] 核心数据结构定义
- [ ] 配置管理系统
- [ ] 基础HTTP服务框架
- [ ] 日志和监控基础设施

### Phase 2: SSP接入 (Week 5-8)
- [ ] SSP适配器框架
- [ ] OpenRTB标准适配器
- [ ] 小米/快手适配器
- [ ] 协议转换和验证

### Phase 3: 核心引擎 (Week 9-12)
- [ ] 用户画像服务集成
- [ ] 流量分发定向
- [ ] DSP适配器框架
- [ ] 竞价广播机制

### Phase 4: 决策引擎 (Week 13-16)
- [ ] 竞价决策算法
- [ ] 定价引擎实现
- [ ] 候选排序逻辑
- [ ] 性能优化

### Phase 5: 完善集成 (Week 17-20)
- [ ] 全链路测试
- [ ] 性能调优
- [ ] 监控告警完善
- [ ] Docker容器化部署

---

## 下一步行动

请确认以上设计文档的整体结构和内容方向是否符合预期，我将根据您的反馈进行相应调整，然后开始详细的实现工作。

需要确认的关键点：
1. 技术选型是否合适？
2. 模块划分和开发顺序是否合理？
3. Docker部署方案是否满足需求？
4. 性能目标（QPS 10000，延迟 ≤ 50ms）是否可行？
5. 还有哪些需要补充或调整的部分？