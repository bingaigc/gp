.PHONY: help build run test clean deps

BINARY_NAME=alpha-detector
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "1.0.0")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

## help: 显示帮助信息
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: 构建二进制文件
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/alpha

## run: 运行应用 (开发模式)
run:
	go run ./cmd/alpha scan --config configs/config.yaml

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
