# ADMUX部署指南

本目录包含ADMUX项目的独立部署配置，支持广告业务要求的分离部署和运维。

## 目录结构

```
deployments/
├── README.md                    # 部署指南
├── adxserver/                   # ADX服务器部署配置
│   ├── build.sh                 # ADX服务器构建脚本
│   ├── bootstrap.sh             # ADX服务器启动脚本
│   └── docker/
│       └── Dockerfile           # ADX服务器Docker文件
├── trackingserver/              # Tracking服务器部署配置
│   ├── build.sh                 # Tracking服务器构建脚本
│   ├── bootstrap.sh             # Tracking服务器启动脚本
│   └── docker/
│       └── Dockerfile           # Tracking服务器Docker文件
└── monitoring/                  # 监控服务配置
    ├── docker-compose.yml       # 监控服务编排
    └── conf/                    # 监控配置文件
        ├── prometheus.yml       # Prometheus配置
        ├── redis.conf           # Redis配置
        └── grafana/             # Grafana配置
            └── provisioning/
                └── datasources/
                    └── prometheus.yml
```

## 配置管理

所有服务都使用 `RUN_TYPE` 环境变量来控制配置文件加载：

- `RUN_TYPE=test`：加载 `conf/test.yaml`
- `RUN_TYPE=prod`：加载 `conf/prod.yaml`

配置文件位置：
- ADX服务器：`cmd/adxserver/conf/`
- Tracking服务器：`cmd/trcking_server/conf/`

## 独立部署

### 1. ADX服务器

#### 构建
```bash
cd deployments/adxserver
./build.sh
```

#### 运行
```bash
# 测试环境
docker run -d --name adx-server \
  -p 8080:8080 \
  -e RUN_TYPE=test \
  admux/adx-server:latest

# 生产环境
docker run -d --name adx-server \
  -p 8080:8080 \
  -e RUN_TYPE=prod \
  -e REDIS_URL=redis://prod-redis:6379 \
  -e LOG_LEVEL=warn \
  admux/adx-server:latest
```

### 2. Tracking服务器

#### 构建
```bash
cd deployments/trackingserver
./build.sh
```

#### 运行
```bash
# 测试环境
docker run -d --name tracking-server \
  -p 8081:8081 \
  -e RUN_TYPE=test \
  admux/tracking-server:latest

# 生产环境
docker run -d --name tracking-server \
  -p 8081:8081 \
  -e RUN_TYPE=prod \
  -e REDIS_URL=redis://prod-redis:6379 \
  -e TRACKING_DB_DSN="user:pass@tcp(db:3306)/tracking" \
  admux/tracking-server:latest
```

### 3. 监控服务

#### 启动
```bash
cd deployments/monitoring
docker-compose up -d
```

#### 访问地址
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)
- Redis: localhost:6379

## 环境变量

### 通用环境变量
- `RUN_TYPE`: 运行类型 (test/prod)
- `CONFIG_PATH`: 配置文件路径 (默认: /app)
- `LOG_PATH`: 日志路径 (默认: /app/logs)

### ADX服务器环境变量
- `REDIS_URL`: Redis连接地址
- `REDIS_PASSWORD`: Redis密码
- `SERVER_PORT`: 服务端口 (默认: 8080)
- `DSP*_AUTH_TOKEN`: DSP认证令牌

### Tracking服务器环境变量
- `REDIS_URL`: Redis连接地址
- `TRACKING_DB_DSN`: 数据库连接字符串
- `SERVER_PORT`: 服务端口 (默认: 8081)
- `POSTBACK_URL`: 回调URL

## 快速启动

### 开发环境
```bash
# 1. 启动监控服务
cd deployments/monitoring
docker-compose up -d

# 2. 构建并启动ADX服务器
cd ../adxserver
./build.sh
docker run -d --name adx-server \
  -p 8080:8080 \
  -e RUN_TYPE=test \
  --network admux-monitoring_admux-monitoring \
  admux/adx-server:latest

# 3. 构建并启动Tracking服务器
cd ../trackingserver
./build.sh
docker run -d --name tracking-server \
  -p 8081:8081 \
  -e RUN_TYPE=test \
  --network admux-monitoring_admux-monitoring \
  admux/tracking-server:latest
```

### 生产环境
```bash
# 1. 设置生产环境变量
export RUN_TYPE=prod
export REDIS_URL=redis://prod-redis:6379
export GRAFANA_PASSWORD=your_secure_password

# 2. 启动监控服务
cd deployments/monitoring
docker-compose up -d

# 3. 构建生产镜像
cd ../adxserver
./build.sh
docker run -d --name adx-server-prod \
  -p 8080:8080 \
  -e RUN_TYPE=prod \
  -e REDIS_URL=$REDIS_URL \
  --restart unless-stopped \
  admux/adx-server:prod

cd ../trackingserver
./build.sh
docker run -d --name tracking-server-prod \
  -p 8081:8081 \
  -e RUN_TYPE=prod \
  -e REDIS_URL=$REDIS_URL \
  -e TRACKING_DB_DSN="prod_user:prod_pass@tcp(prod-db:3306)/admux_tracking" \
  --restart unless-stopped \
  admux/tracking-server:prod
```

## 健康检查

所有服务都提供健康检查端点：

```bash
# ADX服务器
curl http://localhost:8080/health

# Tracking服务器
curl http://localhost:8081/health

# 监控服务
curl http://localhost:9090/-/healthy
curl http://localhost:3000/api/health
```

## 日志查看

```bash
# 查看ADX服务器日志
docker logs -f adx-server

# 查看Tracking服务器日志
docker logs -f tracking-server

# 查看监控服务日志
cd deployments/monitoring
docker-compose logs -f
```

## 监控指标

访问以下地址查看监控指标：

- ADX服务器: http://localhost:8080/metrics
- Tracking服务器: http://localhost:8081/metrics
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000

## 故障排除

### 常见问题

1. **配置文件找不到**
   - 确保 `RUN_TYPE` 环境变量设置正确
   - 检查配置文件路径是否存在

2. **服务无法启动**
   - 检查端口是否被占用
   - 查看容器日志：`docker logs <container_name>`

3. **健康检查失败**
   - 检查服务是否正常启动
   - 验证健康检查端点是否可访问

### 调试命令

```bash
# 进入容器调试
docker exec -it adx-server /bin/sh
docker exec -it tracking-server /bin/sh

# 检查网络连接
docker network ls
docker network inspect <network_name>

# 查看资源使用
docker stats
```

## 安全配置

1. **生产环境密码**
   - 设置强密码：`export GRAFANA_PASSWORD=your_secure_password`
   - 配置Redis密码：`export REDIS_PASSWORD=your_redis_password`

2. **网络安全**
   - 使用内部网络进行服务间通信
   - 只暴露必要的端口

3. **访问控制**
   - 配置防火墙规则
   - 使用HTTPS进行外部通信

## 备份和恢复

```bash
# 备份Redis数据
docker exec admux-redis redis-cli BGSAVE
docker cp admux-redis:/data/dump.rdb ./backup/

# 备份Prometheus数据
docker cp admux-prometheus:/prometheus ./backup/

# 备份Grafana数据
docker cp admux-grafana:/var/lib/grafana ./backup/
```

## 扩展部署

### 水平扩展

```bash
# 扩展ADX服务器实例
docker run -d --name adx-server-2 \
  -p 8082:8080 \
  -e RUN_TYPE=prod \
  admux/adx-server:prod

# 使用负载均衡器
# 配置Nginx或HAProxy进行负载均衡
```

### 垂直扩展

```bash
# 增加资源限制
docker run -d --name adx-server \
  --memory=2g \
  --cpus=1.0 \
  -p 8080:8080 \
  -e RUN_TYPE=prod \
  admux/adx-server:prod
```

## 联系支持

如需技术支持，请联系运维团队或查看项目文档。