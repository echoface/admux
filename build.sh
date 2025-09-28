#!/bin/bash

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 默认配置
VERSION=${VERSION:-dev}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}
REGISTRY=${REGISTRY:-localhost:5000}
IMAGE_NAME=${IMAGE_NAME:-admux/adx-server}
PUSH=${PUSH:-false}

# 显示构建信息
show_build_info() {
    log_info "构建配置:"
    echo "  版本: $VERSION"
    echo "  构建时间: $BUILD_TIME"
    echo "  Git提交: $GIT_COMMIT"
    echo "  镜像名称: $IMAGE_NAME"
    echo "  推送到仓库: $PUSH"
    echo ""
}

# 检查依赖
check_dependencies() {
    log_step "检查构建依赖..."

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

# 清理旧的构建文件
clean_build() {
    log_step "清理旧的构建文件..."

    # 停止并删除容器
    docker-compose down --remove-orphans 2>/dev/null || true

    # 删除旧的镜像
    docker rmi "$IMAGE_NAME:latest" 2>/dev/null || true
    docker rmi "$IMAGE_NAME:$VERSION" 2>/dev/null || true

    # 清理未使用的镜像和容器
    docker image prune -f
    docker volume prune -f

    log_info "清理完成"
}

# 创建必要的目录
setup_directories() {
    log_step "创建必要的目录..."

    mkdir -p logs configs data monitoring/{prometheus,grafana/{dashboards,provisioning}} deployments/{nginx/{conf.d,ssl},scripts}

    log_info "目录创建完成"
}

# 构建Docker镜像
build_image() {
    log_step "构建Docker镜像..."

    # 构建参数
    build_args=(
        "--build-arg" "VERSION=$VERSION"
        "--build-arg" "BUILD_TIME=$BUILD_TIME"
        "--build-arg" "GIT_COMMIT=$GIT_COMMIT"
    )

    # 构建镜像
    docker build "${build_args[@]}" -t "$IMAGE_NAME:$VERSION" -t "$IMAGE_NAME:latest" .

    if [ $? -eq 0 ]; then
        log_info "Docker镜像构建成功"

        # 显示镜像信息
        log_info "镜像信息:"
        docker images "$IMAGE_NAME" --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"
    else
        log_error "Docker镜像构建失败"
        exit 1
    fi
}

# 推送镜像到仓库
push_image() {
    if [ "$PUSH" = "true" ]; then
        log_step "推送镜像到仓库..."

        # 推送版本标签
        docker push "$IMAGE_NAME:$VERSION"

        # 推送latest标签
        docker push "$IMAGE_NAME:latest"

        log_info "镜像推送完成"
    fi
}

# 运行测试
run_tests() {
    log_step "运行测试..."

    # 启动服务
    docker-compose up -d redis adx-server

    # 等待服务启动
    sleep 10

    # 健康检查
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        log_info "服务健康检查通过"
    else
        log_error "服务健康检查失败"
        docker-compose logs adx-server
        exit 1
    fi

    # 运行基础功能测试
    log_info "运行基础功能测试..."

    # 测试根端点
    if curl -f http://localhost:8080/ > /dev/null 2>&1; then
        log_info "根端点测试通过"
    else
        log_error "根端点测试失败"
    fi

    # 测试健康检查端点
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        log_info "健康检查端点测试通过"
    else
        log_error "健康检查端点测试失败"
    fi

    # 测试指标端点
    if curl -f http://localhost:8080/metrics > /dev/null 2>&1; then
        log_info "指标端点测试通过"
    else
        log_error "指标端点测试失败"
    fi

    log_info "测试完成"
}

# 生成构建报告
generate_report() {
    log_step "生成构建报告..."

    report_file="build-report-$(date +%Y%m%d-%H%M%S).txt"

    cat > "$report_file" << EOF
ADMUX ADX服务器构建报告
========================

构建信息:
- 版本: $VERSION
- 构建时间: $BUILD_TIME
- Git提交: $GIT_COMMIT
- 镜像名称: $IMAGE_NAME

镜像信息:
$(docker images "$IMAGE_NAME" --format "{{.Repository}}:{{.Tag}} - {{.Size}} - {{.CreatedAt}}")

容器信息:
$(docker ps --filter "name=admux" --format "{{.Names}} - {{.Status}} - {{.Ports}}")

构建完成时间: $(date)
EOF

    log_info "构建报告已生成: $report_file"
}

# 主函数
main() {
    log_info "开始构建ADMUX ADX服务器..."

    show_build_info
    check_dependencies

    if [ "$1" = "clean" ]; then
        clean_build
        exit 0
    fi

    setup_directories
    clean_build
    build_image
    push_image

    if [ "$1" = "test" ]; then
        run_tests
    fi

    generate_report

    log_info "构建完成！"
    log_info "使用以下命令启动服务:"
    echo "  docker-compose up -d"
    echo ""
    log_info "服务访问地址:"
    echo "  ADX服务: http://localhost:8080"
    echo "  健康检查: http://localhost:8080/health"
    echo "  指标监控: http://localhost:8080/metrics"
    echo ""
    log_info "如需启动监控服务，请运行:"
    echo "  docker-compose --profile monitoring up -d"
}

# 解析命令行参数
case "${1:-}" in
    "clean")
        main clean
        ;;
    "test")
        main test
        ;;
    "help"|"-h"|"--help")
        echo "用法: $0 [选项]"
        echo ""
        echo "选项:"
        echo "  clean    清理构建文件和容器"
        echo "  test     构建后运行测试"
        echo "  help     显示帮助信息"
        echo ""
        echo "环境变量:"
        echo "  VERSION        设置版本标签 (默认: dev)"
        echo "  BUILD_TIME     设置构建时间"
        echo "  GIT_COMMIT     设置Git提交ID"
        echo "  REGISTRY       设置镜像仓库地址"
        echo "  IMAGE_NAME     设置镜像名称"
        echo "  PUSH           是否推送镜像到仓库 (默认: false)"
        exit 0
        ;;
    *)
        main
        ;;
esac