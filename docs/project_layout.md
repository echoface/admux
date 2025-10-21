# ADMUX 项目布局指南文档

## 当前项目结构

```
admux/
├── golang/               # Go 后端服务统一目录
│   ├── cmd/              # 各服务入口文件
│   │   ├── ad_mgr/       # 广告管理服务
│   │   ├── adx_engine/   # ADX 引擎服务
│   │   ├── dsp_engine/   # DSP 引擎服务
│   │   └── media_mgr/    # 媒体管理服务
│   ├── internal/         # 内部包，不对外暴露
│   │   ├── adx_engine/   # ADX 引擎实现
│   │   ├── dsp_engine/   # DSP 引擎实现
│   │   ├── event_tracking/ # 事件追踪服务
│   │   ├── adv_platform/ # 广告平台
│   │   └── media_platform/ # 媒体平台
│   ├── pkg/              # 项目内可复用包
│   │   ├── protogen/     # Protobuf 生成的代码
│   │   ├── utils/        # 工具函数
│   │   └── filehelper/   # 文件处理辅助
│   ├── go.mod
│   ├── go.sum
│   └── Makefile
├── web/                  # 前端应用
│   ├── admin/            # React 管理端
│   ├── customer/         # React 用户端
│   └── shared/           # 前端共享代码
├── mobile/               # 移动端（预留）
│   ├── ios/
│   └── android/
├── shared/               # 跨栈共享代码
│   ├── proto/            # gRPC 协议定义
│   │   ├── admux/        # ADMUX 内部协议
│   │   └── kuaishou/     # 快手 SSP 协议
│   ├── types/            # TypeScript 类型定义
│   └── contracts/        # API 契约文档
├── infrastructure/       # 基础设施代码
│   ├── docker/           # Docker 配置
│   │   └── docker-compose.yml
│   ├── k8s/              # Kubernetes 配置
│   └── terraform/        # Terraform 基础设施代码
├── tools/                # 开发和运维工具
│   ├── benchmark/        # 性能测试工具
│   ├── debugging/        # 调试工具
│   └── migration/        # 数据迁移工具
├── scripts/              # 脚本文件
├── docs/                 # 项目文档
│   └── adx/              # ADX 相关文档
├── java/                 # Java 服务（预留）
├── Makefile              # 项目构建脚本
├── go.work               # Go workspace 配置
└── README.md
```

## 当前设计决策

### 1. Go 服务统一管理
- **原因**: 使用单一 `golang` 目录管理所有 Go 服务，简化依赖管理和构建流程
- **优势**: 统一的 go.mod，减少依赖冲突，便于代码共享和重构
- **结构**: `cmd/` 作为服务入口，`internal/` 作为业务逻辑，`pkg/` 作为公共库

### 2. 协议文件集中管理
- **位置**: `shared/proto/` 目录集中管理所有 protobuf 文件
- **生成**: 通过根目录 Makefile 统一生成到 `golang/pkg/protogen/`
- **包含**: ADMUX 内部协议和第三方 SSP 协议（如快手）

### 3. 前端模块化
- **分离**: 管理端（admin）和用户端（customer）完全分离
- **共享**: 前端公共组件和工具放在 `web/shared/`
- **技术**: 基于 React + TypeScript + Vite

### 4. 基础设施代码化
- **容器化**: Docker 配置标准化
- **编排**: Kubernetes 和 docker-compose 支持
- **云原生**: Terraform 管理云资源

## 命名约定

### 目录命名原则
- ✅ **使用小写字母和下划线**: `adx_engine`, `event_tracking`
- ✅ **使用简写**: `cmd`, `pkg`, `proto`, `k8s`
- ✅ **功能导向**: `media_platform`, `adv_platform`
- ❌ **避免**: 驼峰命名（`AdxEngine`）、中划线（`adx-engine`）

### 文件命名原则
- Go 文件: 使用下划线分隔的小写命名 `adx_server.go`
- 配置文件: 使用下划线分隔 `dev.yaml`, `prod.yaml`
- 文档文件: 使用下划线分隔 `project_layout.md`

## 后续开发应遵循的原则

### 1. 服务拆分原则
```bash
# 新增服务时的目录结构
golang/
├── cmd/
│   └── new_service/
│       └── main.go
├── internal/
│   └── new_service/
│       ├── handler.go
│       ├── service.go
│       └── model.go
└── pkg/
    └── shared_lib/  # 如需跨服务共享
```

### 2. 依赖管理原则
- **内部包**: `internal/` 下的包只对服务内部可见，不可跨服务引用
- **公共包**: `pkg/` 下的包可在所有服务间共享
- **外部依赖**: 在 `go.mod` 中统一管理，避免重复引入

### 3. 接口设计原则
- **协议优先**: 新功能先定义 protobuf 文件，再生成代码
- **版本管理**: API 协议变更时保持向后兼容
- **文档同步**: 重要接口需在 `docs/` 中同步文档

### 4. 配置管理原则
```yaml
# 配置文件命名规范
config/
├── dev.yaml      # 开发环境
├── test.yaml     # 测试环境
├── staging.yaml  # 预发布环境
└── prod.yaml     # 生产环境
```

### 5. 构建和部署原则
- **统一构建**: 使用根目录 Makefile 统一管理构建流程
- **容器化**: 所有服务必须提供 Dockerfile
- **环境隔离**: 不同环境使用不同的配置文件

### 6. 代码组织原则
- **单一职责**: 每个包只负责一个明确的功能域
- **依赖倒置**: 高层模块不依赖低层模块，都依赖抽象
- **接口隔离**: 不要强迫客户端依赖它们不使用的接口

## 开发工作流

### 1. 新功能开发流程
1. 在 `shared/proto/` 中定义协议文件
2. 运行 `make proto` 生成代码
3. 在 `golang/internal/` 中实现业务逻辑
4. 在 `golang/cmd/` 中创建服务入口
5. 更新相关文档

### 2. 协议更新流程
1. 修改 `shared/proto/` 中的 `.proto` 文件
2. 运行 `make proto` 重新生成代码
3. 更新相关的服务实现
4. 运行测试确保兼容性

### 3. 前端开发流程
1. 在 `web/admin/` 或 `web/customer/` 中开发
2. 共享组件放在 `web/shared/`
3. 类型定义放在 `shared/types/`
4. 与后端 API 协议保持同步

## 工具和命令

### 常用命令
```bash
# 生成 protobuf 代码
make proto

# 清理生成的文件
make clean-proto

# 构建所有服务
make build

# 运行测试
make test

# 查看帮助
make help
```

### 目录导航
```bash
# 快速跳转到常用目录
cd golang/cmd/adx_engine    # ADX 引擎入口
cd internal/adx_engine      # ADX 引擎实现
cd shared/proto             # 协议定义
cd web/admin                # 管理端前端
cd infrastructure/docker    # Docker 配置
```

## 最佳实践

### 1. 代码质量
- 所有新代码必须包含单元测试
- 使用 `gofmt` 格式化代码
- 遵循 Go 官方编码规范

### 2. 安全性
- 敏感配置使用环境变量
- API 接口实现认证和授权
- 定期更新依赖包版本

### 3. 性能优化
- 使用连接池管理数据库连接
- 实现缓存策略减少重复计算
- 监控关键性能指标

### 4. 可观测性
- 统一日志格式和级别
- 实现链路追踪
- 添加健康检查端点

## 版本历史
- **v1.0**: 初始布局设计（基于 services 目录）
- **v2.0**: 重构为 golang 统一管理目录（当前版本）
