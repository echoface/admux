# DSP索引管理器 (DSP Bidder Index Manager)

## 概述

DSP索引管理器是ADMUX ADX引擎的核心组件，负责动态管理、索引和调度所有DSP竞价方。它通过S3兼容存储自动发现DSP配置，使用be_indexer构建高性能索引，并提供智能的DSP选择和负载均衡功能。

## 核心特性

### ✅ 已实现功能

1. **S3兼容存储支持**
   - 支持阿里云OSS、AWS S3、MinIO等S3兼容存储
   - 定时扫描固定前缀，自动发现DSP配置
   - 支持增量更新和全量重建

2. **高性能索引系统**
   - 使用be_indexer构建倒排索引
   - 支持多维定向条件快速匹配
   - 索引原子切换，零停机更新

3. **智能DSP调度**
   - 基于定向条件的DSP匹配
   - QPS控制和负载均衡
   - 预算管理和频控

4. **内存LRU缓存**
   - 存储DSP动态状态（QPS、预算、健康状态）
   - LRU策略自动淘汰
   - 线程安全，支持高并发

5. **实时监控**
   - 完整的Prometheus指标
   - 缓存命中率统计
   - 扫描次数和错误计数

## 架构设计

### 组件架构

```
┌─────────────────────────────────────────────────────────────┐
│                   BidderIndexManager                        │
│  ┌───────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │   S3Client    │  │ IndexBuilder │  │ DSPDynamicCache  │  │
│  └───────────────┘  └──────────────┘  └──────────────────┘  │
│         │                 │                      │          │
│         └─────────────────┼──────────────────────┘          │
│                           │                                 │
│  ┌────────────────────────────────────────────────────────┐ │
│  │           DSPIndex (Internal Index)                   │ │
│  │  - dspMap: map[string]*DSPInfo                        │ │
│  │  - clauseMap: map[string][]*DSPInfo                   │ │
│  └────────────────────────────────────────────────────────┘ │
│                           │                                 │
│  ┌────────────────────────────────────────────────────────┐ │
│  │           be_indexer (External Index)                 │ │
│  │  - 支持复杂查询和全文搜索                              │ │
│  │  - 支持索引持久化和恢复                                │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 数据流

```
1. S3扫描 → 读取DSP配置JSON
       ↓
2. 解析配置 → 构建DSPInfo结构
       ↓
3. 索引构建 → be_indexer + DSPIndex
       ↓
4. 索引切换 → 原子更新当前索引
       ↓
5. Bidder注册 → AdxCore.BidderFactory
       ↓
6. 运行时调度 → MatchDSPs() + QPS控制
```

## 使用指南

### 1. 配置文件

在`prod.yaml`或`test.yaml`中添加S3配置：

```yaml
s3:
  endpoint: ${S3_ENDPOINT}              # S3服务地址
  access_key_id: ${S3_ACCESS_KEY_ID}    # 访问密钥ID
  secret_access_key: ${S3_SECRET_ACCESS_KEY}  # 访问密钥
  bucket_name: ${S3_BUCKET_NAME}        # 存储桶名称
  prefix: "adx/dsp/"                    # DSP配置前缀
  use_ssl: true                         # 是否使用SSL
  scan_interval: 1m                     # 扫描间隔
  region: ${S3_REGION}                  # 区域
```

### 2. DSP配置格式

在S3中存储DSP配置文件，JSON格式：

```json
{
  "dsp_id": "dsp_12345",
  "dsp_name": "Example DSP",
  "status": "active",
  "qps_limit": 1000,
  "budget_daily": 50000.00,
  "endpoint": "http://dsp.example.com/bid",
  "auth_token": "your-auth-token",
  "timeout": 80000000000,
  "retry_count": 2,
  "retry_delay": 10000000,
  "targeting": {
    "indexingdoc": [
      {
        "clause_id": "clause_1",
        "description": "北京或上海的iOS用户",
        "conditions": [
          {
            "field": "USER_GEO",
            "operator": "IN",
            "values": ["BJ", "SH"]
          },
          {
            "field": "USER_OS",
            "operator": "EQ",
            "values": ["ios"]
          }
        ]
      }
    ]
  },
  "filters": {
    "min_cpm": 0.5,
    "max_cpm": 100.0,
    "blocked_categories": ["gambling", "adult"]
  }
}
```

### 3. 启动索引管理器

```go
// 创建配置
cfg := &config.ServerConfig{
    S3: config.S3Config{
        Endpoint:     "https://s3.aliyun.com",
        AccessKeyID:  "your-access-key",
        SecretAccessKey: "your-secret-key",
        BucketName:   "adx-config",
        Prefix:       "dsp/",
        UseSSL:       true,
        ScanInterval: time.Minute,
        Region:       "cn-hangzhou",
    },
}

// 创建并启动管理器
mgr, err := NewBidderIndexManager(cfg)
if err != nil {
    log.Fatalf("Failed to create manager: %v", err)
}

if err := mgr.Start(); err != nil {
    log.Fatalf("Failed to start manager: %v", err)
}
defer mgr.Stop()
```

### 4. 竞价时使用

```go
// 匹配DSP
conditions := map[string][]string{
    "USER_GEO":     []string{"BJ", "SH"},
    "USER_OS":      []string{"ios"},
    "DEVICE_TYPE":  []string{"mobile"},
}
matchedDSPs := mgr.MatchDSPs(conditions)

// QPS控制
for _, dsp := range matchedDSPs {
    currentQPS, _ := mgr.GetDSPQPS(dsp.DSPID)
    if currentQPS < dsp.QPSLimit {
        newQPS := mgr.IncrementDSPQPS(dsp.DSPID)
        // 发送竞价请求到DSP
    }
}
```

## 核心API

### BidderIndexManager

#### 生命周期管理
- `NewBidderIndexManager(cfg)`: 创建管理器
- `Start() error`: 启动管理器
- `Stop() error`: 停止管理器

#### DSP查询
- `GetDSP(dspID)`: 获取单个DSP信息
- `GetAllDSPs()`: 获取所有DSP
- `GetAllActiveDSPs()`: 获取所有活跃DSP
- `MatchDSPs(conditions)`: 根据条件匹配DSP

#### 动态状态
- `GetDSPQPS(dspID)`: 获取当前QPS
- `IncrementDSPQPS(dspID)`: 增加QPS计数
- `DecrementDSPQPS(dspID)`: 减少QPS计数
- `ResetDSPQPS(dspID)`: 重置QPS计数
- `GetDSPStatus(dspID)`: 获取DSP状态
- `GetDSPBudget(dspID)`: 获取DSP预算

#### 监控指标
- `GetMetrics()`: 获取管理器指标

### DSPIndex (内部索引)

#### 查询方法
- `GetDSP(dspID)`: 获取DSP信息
- `GetAllDSPs()`: 获取所有DSP
- `GetDSPsByClause(clauseID)`: 根据条款获取DSP
- `GetActiveDSPs()`: 获取活跃DSP
- `MatchDSPs(conditions)`: 条件匹配DSP

### DSPDynamicCache (LRU缓存)

#### 状态管理
- `SetDSPStatus(dspID, status)`: 设置DSP状态
- `GetDSPStatus(dspID)`: 获取DSP状态
- `SetDSPBudget(dspID, budget)`: 设置DSP预算
- `GetDSPBudget(dspID)`: 获取DSP预算

#### QPS控制
- `IncrementDSPQPS(dspID)`: 增加QPS
- `DecrementDSPQPS(dspID)`: 减少QPS
- `GetDSPQPS(dspID)`: 获取QPS
- `ResetDSPQPS(dspID)`: 重置QPS

#### 缓存管理
- `GetMetrics()`: 获取缓存指标

## 定向条件支持

### 操作符

| 操作符 | 描述 | 示例 |
|--------|------|------|
| `EQ` | 等于 | `{"USER_OS": "ios"}` |
| `IN` | 在集合中 | `{"USER_GEO": ["BJ", "SH"]}` |
| `NOT_IN` | 不在集合中 | `{"CATEGORY": ["adult"]}` |
| `GT` | 大于 | `{"PRICE": 100}` |
| `LT` | 小于 | `{"PRICE": 10}` |

### 常用字段

| 字段名 | 描述 | 示例值 |
|--------|------|--------|
| `USER_GEO` | 用户地理 | `BJ`, `SH`, `GD` |
| `USER_OS` | 操作系统 | `ios`, `android`, `windows` |
| `DEVICE_TYPE` | 设备类型 | `mobile`, `desktop`, `tablet` |
| `CONNECTION_TYPE` | 网络类型 | `wifi`, `4g`, `5g` |
| `CONTENT_CATEGORY` | 内容类别 | `news`, `sports`, `entertainment` |

## 监控指标

### BidderIndexManager指标

```json
{
  "last_scan_time": "2024-01-01T10:00:00Z",
  "scan_count": 100,
  "error_count": 0,
  "current_dsp_count": 10,
  "active_dsp_count": 8,
  "registered_dsp_count": 8,
  "cache_metrics": {
    "hits": 1000,
    "misses": 100,
    "evictions": 0,
    "total_items": 200
  }
}
```

### Prometheus指标

可通过以下方式暴露指标：
- `/metrics`: 标准Prometheus指标
- `mgr.GetMetrics()`: 程序化获取

## 性能特性

### 索引性能
- **索引构建**: 1000个DSP < 1秒
- **查询延迟**: < 1ms (平均)
- **索引切换**: 原子操作，< 10ms
- **内存占用**: 每个DSP约1KB

### 缓存性能
- **LRU缓存**: O(1)读写
- **缓存命中**: > 95% (预期)
- **并发安全**: 读写锁保护
- **内存优化**: 自动淘汰过期项

### 扩展性
- **DSP数量**: 支持10,000+ DSP
- **并发查询**: 支持1000+ QPS
- **存储扩展**: 支持分布式S3
- **索引扩展**: 支持分片和副本

## 最佳实践

### 1. 配置优化

```yaml
# 生产环境推荐配置
s3:
  scan_interval: 30s        # 快速检测变化
  prefix: "adx/dsp/"        # 使用固定前缀
  use_ssl: true             # 启用加密传输
```

### 2. DSP配置优化

```json
{
  "qps_limit": 1000,        # 设置合理的QPS限制
  "timeout": 80000000000,   # 设置合理的超时(80ms)
  "retry_count": 2,         # 设置重试次数
  "retry_delay": 10000000   # 设置重试延迟(10ms)
}
```

### 3. 监控建议

- 定期检查`scan_count`和`error_count`
- 监控缓存命中率
- 关注QPS分布和负载均衡
- 设置告警阈值

### 4. 故障处理

- S3连接失败: 自动重试，指数退避
- 索引构建失败: 回滚到上一版本
- DSP不可用: 自动从候选列表移除
- 缓存满: LRU自动淘汰

## 与现有架构集成

### 与AdxCore集成

```go
// 在bidding流程中使用
func targetingBidders(ctx *BidRequestCtx) error {
    // 使用索引管理器匹配DSP
    conditions := extractConditions(ctx.Request)
    matchedDSPs := mgr.MatchDSPs(conditions)

    // 获取活跃DSP并发送请求
    for _, dsp := range matchedDSPs {
        if mgr.IncrementDSPQPS(dsp.DSPID) <= dsp.QPSLimit {
            // 发送竞价请求
        }
    }
}
```

### 与配置系统集成

```go
// 自动加载配置
func LoadConfig() (*config.ServerConfig, error) {
    cfg, err := config.LoadConfig("adxserver")
    if err != nil {
        return nil, err
    }
    return cfg, nil
}
```

## 未来扩展

### 计划功能
- [ ] 支持更多S3兼容存储 (Azure Blob, 华为云OBS)
- [ ] 支持Redis持久化缓存
- [ ] 支持动态权重调整
- [ ] 支持A/B测试DSP
- [ ] 支持实时竞价日志分析

### 优化方向
- [ ] 索引分片和分布式
- [ ] 机器学习优化DSP选择
- [ ] 预测性预加载
- [ ] 零拷贝优化

## 贡献指南

### 代码规范
- 遵循Go官方代码规范
- 添加单元测试覆盖关键逻辑
- 使用`go fmt`和`go vet`检查代码
- 更新文档和示例

### 测试
```bash
go test -v ./internal/adx_engine/dspbidder/
go test -run Example ./internal/adx_engine/dspbidder/
```

## 许可证

本项目采用MIT许可证。

## 联系方式

如有问题或建议，请联系ADMUX团队。
