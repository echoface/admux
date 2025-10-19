下面本repo项目layout的指南文档， 整体的目录参考如下
```
admux/
├── services/             # 后端服务Go/Java/Node/Python
│   ├── adx_engine/
│   ├──────cmd/
│   ├────────main.go
│   ├────────conf/
│   ├───────────test.yaml
│   ├───────────prod.yaml
│   ├──────handler.go
│   ├──────internal/
│   ├──────pkg/      # 项目相关无需共享的pkg
│   ├──────go.mod
│   ├──────go.sum
│   ├──────Makefile
│   ├──────README.md
│   ├── dsp_engine/
│   ├──────/handler.go
│   ├── ad_tracking/
│   ├── pkg/             # Go后段服务共享的pkg
│   ├── go.work         # gowork 项目目录
├── web/                 # 前端应用
│   ├── admin/           # React 管理端
│   ├── customer/        # React 用户端
│   └── shared/          # 前端共享代码
├── mobile/              # 移动端
│   ├── ios/
│   └── android/
├── shared/              # 跨栈共享
│   ├── proto/           # gRPC 协议
│   ├── types/           # TypeScript 类型
│   └── contracts/       # API 契约
└── docs/                # 文档中心
└────  adx/              # adx项目文档中心
└── infrastructure/      # 基础设施
    ├── k8s/
    ├── docker/
    └── terraform/
```

同时需要遵循如下原则
- 目录名使用简写或下划线命名
