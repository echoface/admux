# ADMUX ADX服务器 Makefile

.PHONY: help build build-adx build-tracking push clean test run run-adx run-tracking stop logs shell lint format check deps docker-build docker-run docker-stop docker-clean docker-logs proto

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
VERSION ?= dev
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
IMAGE_NAME ?= admux/adx-server
REGISTRY ?= localhost:5000

# Go相关变量
GOOS ?= linux
GOARCH ?= amd64
CGO_ENABLED ?= 0

# Docker相关变量
DOCKER_COMPOSE = docker-compose
DOCKER_BUILDKIT ?= 1

# 颜色定义
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m

help: ## 显示帮助信息
	@echo "$(GREEN)ADMUX ADX服务器构建工具$(NC)"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 原有命令保持兼容
proto: ## 生成protobuf代码
	@echo "$(GREEN)生成Go代码从protobuf...$(NC)"
	mkdir -p api/gen
	PATH="$(shell go env GOPATH)/bin:$$PATH" protoc --proto_path=. \
		--go_out=api/gen \
		api/idl/*.proto

# 本地构建
build: proto build-adx build-tracking ## 构建所有二进制文件

build-adx: ## 构建ADX服务器
	@echo "$(GREEN)构建ADX服务器...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags="-w -s -extldflags '-static' -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)" \
		-a -installsuffix cgo \
		-o bin/adx_server ./cmd/adx_server

build-tracking: ## 构建Tracking服务器
	@echo "$(GREEN)构建Tracking服务器...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags="-w -s -extldflags '-static'" \
		-a -installsuffix cgo \
		-o bin/trcking_server ./cmd/trcking_server

# 测试
test: ## 运行测试
	@echo "$(GREEN)运行测试...$(NC)"
	go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)测试完成，覆盖率报告: coverage.out$(NC)"

test-coverage: test ## 生成HTML覆盖率报告
	@echo "$(GREEN)生成HTML覆盖率报告...$(NC)"
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)覆盖率报告已生成: coverage.html$(NC)"

# 代码质量
lint: ## 运行代码检查
	@echo "$(GREEN)运行代码检查...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint未安装，跳过代码检查$(NC)"; \
		echo "安装命令: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

format: ## 格式化代码
	@echo "$(GREEN)格式化代码...$(NC)"
	go fmt ./...
	goimports -w .

check: lint test ## 运行所有检查（代码检查+测试）

deps: ## 更新依赖
	@echo "$(GREEN)更新Go模块依赖...$(NC)"
	go mod tidy
	go mod verify
	go mod download

# 本地运行（保持原有兼容性）
run-adx: build-adx ## 运行ADX服务器
	@echo "$(GREEN)启动ADX服务器...$(NC)"
	./bin/adx_server

run-tracking: build-tracking ## 运行Tracking服务器
	@echo "$(GREEN)启动Tracking服务器...$(NC)"
	./bin/trcking_server

run: ## 同时运行ADX和Tracking服务器
	@echo "$(GREEN)启动所有服务...$(NC)"
	./bin/adx_server &
	./bin/trcking_server

# Docker相关命令
docker-build: ## 构建Docker镜像
	@echo "$(GREEN)构建Docker镜像...$(NC)"
	DOCKER_BUILDKIT=$(DOCKER_BUILDKIT) docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(IMAGE_NAME):$(VERSION) \
		-t $(IMAGE_NAME):latest .
	@echo "$(GREEN)Docker镜像构建完成$(NC)"

docker-run: ## 运行Docker容器
	@echo "$(GREEN)启动Docker服务...$(NC)"
	$(DOCKER_COMPOSE) up -d

docker-stop: ## 停止Docker容器
	@echo "$(GREEN)停止Docker服务...$(NC)"
	$(DOCKER_COMPOSE) down

docker-clean: ## 清理Docker资源
	@echo "$(GREEN)清理Docker资源...$(NC)"
	$(DOCKER_COMPOSE) down -v --remove-orphans
	docker system prune -f
	docker volume prune -f

docker-logs: ## 查看Docker日志
	$(DOCKER_COMPOSE) logs -f

docker-shell: ## 进入ADX服务器容器
	docker exec -it admux-adx-server /bin/sh

# 监控相关
monitoring-up: ## 启动监控服务
	@echo "$(GREEN)启动监控服务...$(NC)"
	$(DOCKER_COMPOSE) up -d prometheus grafana

monitoring-down: ## 停止监控服务
	@echo "$(GREEN)停止监控服务...$(NC)"
	$(DOCKER_COMPOSE) stop prometheus grafana

# 生产部署相关
build-prod: ## 生产环境构建
	$(MAKE) build VERSION=production GOOS=linux GOARCH=amd64

docker-build-prod: ## 生产环境Docker构建
	$(MAKE) docker-build VERSION=production

push: ## 推送Docker镜像
	@echo "$(GREEN)推送Docker镜像...$(NC)"
	docker push $(IMAGE_NAME):$(VERSION)
	docker push $(IMAGE_NAME):latest

# 清理命令（保持原有兼容性）
clean: ## 清理构建文件
	@echo "$(GREEN)清理构建文件...$(NC)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -rf api/gen/go/
	$(MAKE) docker-clean

# 工具命令
install-tools: ## 安装开发工具
	@echo "$(GREEN)安装开发工具...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/air-verse/air@latest

dev: ## 开发模式（热重载）
	@echo "$(GREEN)启动开发模式...$(NC)"
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "$(YELLOW)air未安装，使用普通运行模式$(NC)"; \
		go run ./cmd/adx_server; \
	fi

# 健康检查
health: ## 检查服务健康状态
	@echo "$(GREEN)检查服务健康状态...$(NC)"
	@curl -f http://localhost:8080/health 2>/dev/null && echo "$(GREEN)✓ ADX服务器健康$(NC)" || echo "$(RED)✗ ADX服务器不健康$(NC)"
	@curl -f http://localhost:8080/metrics 2>/dev/null && echo "$(GREEN)✓ 指标端点正常$(NC)" || echo "$(RED)✗ 指标端点异常$(NC)"

# 版本信息
version: ## 显示版本信息
	@echo "$(GREEN)ADMUX ADX服务器版本信息$(NC)"
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Git提交: $(GIT_COMMIT)"
	@echo "Go版本: $(shell go version)"

# 快速启动命令
quick-start: ## 快速启动（构建+运行）
	@echo "$(GREEN)快速启动ADMUX ADX服务器...$(NC)"
	$(MAKE) deps
	$(MAKE) docker-build
	$(MAKE) docker-run
	@echo "$(GREEN)等待服务启动...$(NC)"
	sleep 10
	$(MAKE) health
	@echo "$(GREEN)服务已启动！$(NC)"
	@echo "访问地址:"
	@echo "  ADX服务: http://localhost:8080"
	@echo "  健康检查: http://localhost:8080/health"
	@echo "  指标监控: http://localhost:8080/metrics"
