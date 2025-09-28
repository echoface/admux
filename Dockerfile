# 多阶段构建 - 构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要工具
RUN apk add --no-cache git ca-certificates tzdata curl

# 设置时区
ENV TZ=Asia/Shanghai

# 复制go mod文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# 复制源代码
COPY . .

# 构建参数
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

# 如果未提供构建时间，使用当前时间
RUN if [ -z "$BUILD_TIME" ]; then \
        BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S'); \
    fi

# 如果未提供Git提交ID，获取当前提交
RUN if [ -z "$GIT_COMMIT" ]; then \
        GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
    fi

# 构建ADX服务器二进制文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -extldflags '-static' \
    -X main.Version=${VERSION} \
    -X main.BuildTime=${BUILD_TIME} \
    -X main.GitCommit=${GIT_COMMIT}" \
    -a -installsuffix cgo \
    -o adx-server ./cmd/adx_server

# 构建tracking服务器二进制文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -extldflags '-static'" \
    -a -installsuffix cgo \
    -o tracking-server ./cmd/trcking_server

# 验证二进制文件
RUN ./adx-server --version || echo "Version flag not implemented, continuing..."

# 运行阶段 - ADX服务器
FROM alpine:3.19

# 安装运行时依赖
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    dumb-init \
    && rm -rf /var/cache/apk/*

# 设置时区
ENV TZ=Asia/Shanghai

# 创建应用用户
RUN addgroup -g 1001 -S adx && \
    adduser -u 1001 -S adx -G adx

# 创建必要的目录
RUN mkdir -p /app/configs /app/logs /app/data && \
    chown -R adx:adx /app

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/adx-server .
COPY --from=builder /app/tracking-server .

# 复制配置文件模板
COPY --chown=adx:adx configs/ ./configs/

# 复制启动脚本
COPY --chown=adx:adx deployments/scripts/docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# 创建日志目录
RUN mkdir -p /app/logs && chown -R adx:adx /app/logs

# 切换到非root用户
USER adx

# 暴露端口
EXPOSE 8080 8081

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 设置环境变量
ENV LOG_LEVEL=info
ENV SERVER_PORT=8080
ENV TRACKING_SERVER_PORT=8081
ENV CONFIG_PATH=/app/configs
ENV LOG_PATH=/app/logs

# 使用dumb-init作为PID 1
ENTRYPOINT ["dumb-init", "--"]

# 启动命令
CMD ["/usr/local/bin/docker-entrypoint.sh"]