.PHONY: help build build-cli build-service build-all run test clean deps

BINARY_NAME=alpha-detector
SERVICE_NAME=alpha-detector-service
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "2.0.0")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

## help: 显示帮助信息
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: 构建CLI二进制文件 (默认)
build: build-cli

## build-cli: 构建CLI二进制文件
build-cli:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/alpha

## build-service: 构建Service二进制文件
build-service:
	@echo "Building $(SERVICE_NAME)..."
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(SERVICE_NAME) ./cmd/service

## build-all: 构建所有二进制文件
build-all: build-cli build-service
	@echo "All binaries built successfully"

## run: 运行CLI应用 (开发模式)
run:
	go run ./cmd/alpha scan --config configs/config.yaml

## run-service: 运行Service (开发模式)
run-service:
	RUN_MODE=all LOG_FORMAT=pretty go run ./cmd/service

## docker-build: 构建Docker镜像
docker-build:
	docker build -f Dockerfile.service -t $(SERVICE_NAME):$(VERSION) .

## docker-run: 运行Docker容器
docker-run:
	docker run -d \
		--name alpha-service \
		-p 8080:8080 \
		-p 9090:9090 \
		-e RUN_MODE=all \
		-e LOG_LEVEL=info \
		$(SERVICE_NAME):$(VERSION)

## test: 运行所有测试
test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## test-unit: 运行单元测试
test-unit:
	go test -v -short ./...

## clean: 清理构建产物
clean:
	rm -rf bin/ coverage.out coverage.html
	go clean

## deps: 下载依赖
deps:
	go mod download
	go mod verify

## tidy: 整理依赖
tidy:
	go mod tidy

## fmt: 格式化代码
fmt:
	go fmt ./...

## vet: 代码检查
vet:
	go vet ./...

.DEFAULT_GOAL := help
