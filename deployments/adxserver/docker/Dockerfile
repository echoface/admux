# ADX服务器 Dockerfile
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# 复制源代码
COPY . .

# 构建ADX服务器
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -extldflags '-static'" \
    -a -installsuffix cgo \
    -o adx-server ./cmd/adx_server

# 运行阶段
FROM alpine:3.19

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata curl dumb-init

# 创建应用用户
RUN addgroup -g 1001 -S adx && \
    adduser -u 1001 -S adx -G adx

# 创建必要目录
RUN mkdir -p /app/logs /app/conf && \
    chown -R adx:adx /app

# 复制二进制文件
COPY --from=builder /app/adx-server /app/adx-server

# 复制配置文件
COPY --chown=adx:adx cmd/adx_server/conf /app/conf

# 复制启动脚本
COPY --chown=adx:adx deployments/adxserver/bootstrap.sh /app/bootstrap.sh
RUN chmod +x /app/bootstrap.sh

# 切换到非root用户
USER adx

# 工作目录
WORKDIR /app

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 环境变量
ENV RUN_TYPE=test
ENV CONFIG_PATH=/app

# 启动命令
ENTRYPOINT ["dumb-init", "--"]
CMD ["/app/bootstrap.sh"]