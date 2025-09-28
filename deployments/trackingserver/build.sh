#!/bin/bash

# Tracking服务器构建脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 变量定义
VERSION=${VERSION:-dev}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}
IMAGE_NAME=${IMAGE_NAME:-admux/tracking-server}
REGISTRY=${REGISTRY:-localhost:5000}

# 显示构建信息
show_build_info() {
    log_info "Tracking服务器构建信息:"
    echo "  版本: $VERSION"
    echo "  构建时间: $BUILD_TIME"
    echo "  Git提交: $GIT_COMMIT"
    echo "  镜像名称: $IMAGE_NAME"
    echo ""
}

# 构建Docker镜像
build_image() {
    log_info "构建Tracking服务器Docker镜像..."

    cd "$(dirname "$0")/.."

    docker build \
        -f docker/Dockerfile \
        --build-arg VERSION="$VERSION" \
        --build-arg BUILD_TIME="$BUILD_TIME" \
        --build-arg GIT_COMMIT="$GIT_COMMIT" \
        -t "$IMAGE_NAME:$VERSION" \
        -t "$IMAGE_NAME:latest" \
        .

    if [ $? -eq 0 ]; then
        log_info "Tracking服务器镜像构建成功"
        docker images "$IMAGE_NAME" --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"
    else
        log_error "Tracking服务器镜像构建失败"
        exit 1
    fi
}

# 推送镜像
push_image() {
    if [ "$1" = "push" ]; then
        log_info "推送Tracking服务器镜像..."
        docker push "$IMAGE_NAME:$VERSION"
        docker push "$IMAGE_NAME:latest"
        log_info "镜像推送完成"
    fi
}

# 主函数
main() {
    show_build_info
    build_image
    push_image "$@"

    log_info "Tracking服务器构建完成！"
    log_info "运行命令:"
    echo "  docker run -d --name tracking-server -p 8081:8081 -e RUN_TYPE=test $IMAGE_NAME:latest"
}

main "$@"