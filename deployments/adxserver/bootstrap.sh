#!/bin/sh

# ADX服务器启动脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# 显示启动信息
show_startup_info() {
    log_info "=== ADMUX ADX服务器启动 ==="
    log_info "运行环境: ${RUN_TYPE:-test}"
    log_info "配置路径: ${CONFIG_PATH:-/app}"
    log_info "日志路径: ${LOG_PATH:-/app/logs}"
    log_info "服务端口: ${SERVER_PORT:-8080}"
    log_info "============================="
}

# 检查环境变量
check_env() {
    # 设置默认值
    export RUN_TYPE=${RUN_TYPE:-test}
    export CONFIG_PATH=${CONFIG_PATH:-/app}
    export LOG_PATH=${LOG_PATH:-/app/logs}
    export SERVER_PORT=${SERVER_PORT:-8080}

    # 验证RUN_TYPE
    if [ "$RUN_TYPE" != "test" ] && [ "$RUN_TYPE" != "prod" ]; then
        log_error "无效的RUN_TYPE: $RUN_TYPE，必须是 'test' 或 'prod'"
        exit 1
    fi

    # 创建日志目录
    mkdir -p "$LOG_PATH"
}

# 等待依赖服务
wait_for_dependencies() {
    if [ -n "$REDIS_URL" ]; then
        redis_host=$(echo "$REDIS_URL" | sed 's|redis://||' | cut -d: -f1)
        redis_port=$(echo "$REDIS_URL" | sed 's|redis://||' | cut -d: -f2)

        log_info "等待Redis服务启动: ${redis_host}:${redis_port}"
        timeout 30 sh -c "until nc -z '$redis_host' '$redis_port'; do sleep 1; done"
        log_info "Redis服务已就绪"
    fi
}

# 启动ADX服务器
start_server() {
    log_info "启动ADX服务器..."
    exec /app/adx-server
}

# 主函数
main() {
    show_startup_info
    check_env
    wait_for_dependencies
    start_server
}

main "$@"