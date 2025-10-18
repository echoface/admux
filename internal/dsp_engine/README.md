DSP竞价索引系统需求文档
1. 系统概述
1.1 项目背景
构建一个基于对象存储的DSP竞价索引系统，通过定时任务管理广告数据，利用be_indexer实现高效的布尔表达式索引，支持实时竞价请求。

1.2 核心组件
广告管理平台: 管理广告数据，生成索引文件

对象存储(MinIO): 存储索引文件和版本信息

DSP Engine: 加载索引，处理竞价请求

2. 系统架构
2.1 整体架构图
```
graph TB
    AMP[广告管理平台] -->|上传索引文件| MINIO[MinIO对象存储]
    AMP -->|发布版本| ETCD[版本注册中心]
    DSP[DSP Engine] -->|监听版本| ETCD
    DSP -->|下载索引| MINIO
    BID[竞价请求] --> DSP
    DSP -->|返回广告| BID

    subgraph "数据流"
        AMP -->|生成| IDX[索引文件]
        IDX -->|上传| MINIO
        MINIO -->|下载| DSP
    end
```

广告管理平台 → 对象存储(MinIO) → DSP Engine → 竞价请求
     ↓              ↓               ↓
  生成索引      存储索引文件      加载索引
  发布版本      存储版本信息      检查版本

2.2 数据流向
广告管理平台处理在线广告数据

生成索引文件并上传至MinIO

发布新

我们计划通过对象存储eg： MinIO 来存储； 广告管理平台定时任务将所有在线的广告按照dsp engine的格式要求带版本上传至对象存储中；并附带发布一个最新的版本号； dsp engine 检查这个版本号； 如果发现新的版本，则重新构造索引来支持竞价请求；
将上面的要求结合你的分析整理成一个markdown 格式的架构设计文档；并给出广告定向索引的格式定义设计；以及其他必要的内容； 一些没有明确的内容可以给我列出章节标题；后续讨论补充
DSP竞价索引系统架构设计文档
1. 系统概述
1.1 项目背景
构建基于对象存储的分布式DSP竞价索引系统，通过版本化索引管理实现广告数据的实时更新和高效检索。

1.2 设计目标
解耦: 广告管理平台与DSP Engine解耦

高性能: 支持毫秒级广告检索

可扩展: 支持水平扩展的索引服务

容错性: 优雅的版本切换和回退机制

2. 系统架构
2.1 整体架构图















2.2 核心组件职责
2.2.1 广告管理平台 (Ad Management Platform)
管理广告活动生命周期

定时生成索引文件

上传文件至对象存储

发布版本信息

2.2.2 对象存储 (MinIO)
存储索引数据文件

提供高可用文件服务

版本化文件管理

2.2.3 DSP Engine
监听版本变化

下载并加载索引文件

处理实时竞价请求

索引热更新

3. 数据流设计
3.1 索引生成流程
go
// 伪代码流程
1. 查询所有在线广告活动
2. 转换为be_indexer兼容格式
3. 生成索引文件（压缩格式）
4. 计算文件哈希（完整性校验）
5. 上传至MinIO
6. 发布新版本到注册中心
3.2 索引加载流程
go
// 伪代码流程
1. 监听版本注册中心变化
2. 发现新版本后下载索引文件
3. 验证文件完整性
4. 加载到内存索引器
5. 原子切换至新索引
6. 清理旧版本文件
4. 索引格式定义
4.1 广告定向数据格式
json
{
  "version": "1.0",
  "format": "be_indexer_v1",
  "campaigns": [
    {
      "id": 1001,
      "name": "夏季促销活动",
      "budget": 50000.0,
      "max_bid": 2.5,
      "targeting": {
        "logical_op": "AND",
        "conditions": [
          {
            "field": "age",
            "operator": "BETWEEN",
            "values": [18, 35]
          },
          {
            "field": "gender",
            "operator": "IN",
            "values": ["male"]
          },
          {
            "field": "location",
            "operator": "IN",
            "values": ["beijing", "shanghai", "guangzhou"]
          },
          {
            "logical_op": "OR",
            "conditions": [
              {
                "field": "interest",
                "operator": "IN",
                "values": ["sports", "fitness"]
              },
              {
                "field": "behavior",
                "operator": "IN",
                "values": ["recent_shopper"]
              }
            ]
          }
        ]
      }
    }
  ]
}
4.2 be_indexer 表达式映射
go
// 条件类型定义
type Condition struct {
    Field    string      `json:"field"`
    Operator string      `json:"operator"` // IN, NOT_IN, BETWEEN, GT, LT, EQ
    Values   []interface{} `json:"values"`
}

// 逻辑表达式
type LogicalExpression struct {
    Operator   string        `json:"logical_op"` // AND, OR, NOT
    Conditions []interface{} `json:"conditions"` // Condition 或 LogicalExpression
}
4.3 索引文件格式
4.3.1 主索引文件
text
索引文件命名: indices/{version}/campaigns.{timestamp}.bin
文件格式:
- 文件头: 魔数 + 版本 + 创建时间
- 索引数据: be_indexer 序列化数据
- 文件尾: 校验和
4.3.2 元数据文件
json
{
  "version": "20240115_120000_v1",
  "format_version": "1.0",
  "created_at": "2024-01-15T12:00:00Z",
  "campaign_count": 1500,
  "file_size": 1048576,
  "file_hash": "sha256:abc123...",
  "download_url": "http://minio.example.com/indices/20240115_120000_v1/campaigns.bin"
}
5. 版本管理设计
5.1 版本号规范
text
格式: {YYYYMMDD}_{HHMMSS}_{序列号}
示例: 20240115_120000_v1
5.2 版本注册中心
5.2.1 版本信息结构
json
{
  "current_version": "20240115_120000_v1",
  "previous_version": "20240115_110000_v1",
  "publish_time": "2024-01-15T12:00:00Z",
  "status": "active",
  "metadata_url": "http://minio.example.com/versions/20240115_120000_v1/metadata.json"
}
5.2.2 版本状态流转
text
生成中 → 就绪 → 激活 → 归档
           ↓
         失效
6. DSP Engine 详细设计
6.1 索引管理器
go
type IndexManager struct {
    currentIndex *be_indexer.BooleanIndexer
    nextIndex    *be_indexer.BooleanIndexer
    version      string
    config       *IndexConfig
    minioClient  *minio.Client
    etcdClient   *etcd.Client
}

// 热更新流程
func (m *IndexManager) HotReload(newVersion string) error {
    // 1. 下载新索引文件
    // 2. 验证完整性
    // 3. 加载到nextIndex
    // 4. 原子切换currentIndex
    // 5. 清理旧索引
}
6.2 健康检查机制
go
type HealthStatus struct {
    Version         string    `json:"version"`
    IndexSize       int       `json:"index_size"`
    LastUpdate      time.Time `json:"last_update"`
    MemoryUsage     int64     `json:"memory_usage"`
    QueryLatency    float64   `json:"query_latency_ms"`
    Status          string    `json:"status"` // healthy, degraded, failed
}
7. 存储设计
7.1 MinIO Bucket 结构
text
ad-indices/
├── versions/
│   ├── current -> 20240115_120000_v1
│   └── 20240115_120000_v1/
│       ├── metadata.json
│       └── campaigns.bin
├── archives/
│   └── 20240115_110000_v1/
└── tmp/
    └── upload_xxx.tmp
7.2 文件保留策略
当前版本: 保留1个

历史版本: 保留最近3个

归档版本: 按日期归档，保留30天

8. 监控与告警
8.1 关键指标
索引构建时间

索引文件大小

内存使用量

查询延迟P95/P99

版本切换成功率

8.2 告警规则
索引构建失败

版本切换超时

内存使用超过阈值

查询延迟异常

9. 容错与降级
9.1 故障场景处理
索引文件损坏：回退到上一版本

MinIO不可用：使用本地缓存

版本注册中心故障：本地版本检测

内存不足：优雅降级服务

10. 性能优化
10.1 索引优化
索引分片策略

压缩算法选择

预加载优化

缓存策略

11. 安全考虑
11.1 访问控制
MinIO访问权限

版本信息加密

传输加密(TLS)

12. 部署方案
12.1 环境配置
开发环境配置

测试环境配置

生产环境配置

13. 后续讨论事项
13.1 待明确的技术细节
索引分片策略和粒度

压缩算法选型（Snappy vs Zstd）

版本冲突解决机制

分布式锁实现方案

索引预热策略

13.2 业务规则待确认
广告优先级处理逻辑

预算控制策略

频次控制实现方案

黑白名单管理

13.3 运维相关
日志收集和分析方案

性能基准测试标准

容量规划指南

灾难恢复流程

14. 附录
14.1 配置文件示例
yaml
index:
  rebuild_interval: "5m"
  max_memory_mb: 2048
  compression: "snappy"

minio:
  endpoint: "minio.example.com:9000"
  access_key: "${MINIO_ACCESS_KEY}"
  secret_key: "${MINIO_SECRET_KEY}"
  bucket: "ad-indices"

etcd:
  endpoints: ["etcd1:2379", "etcd2:2379", "etcd3:2379"]
  prefix: "/dsp/indices"
14.2 API接口定义
版本查询接口

强制重建接口

健康检查接口

统计信息接口

文档版本: 1.0
*最后更新: 2024-01-15*
作者: 架构团队
