.PHONY: all help tidy fmt build test test-race test-race-pkg build-race-pkg verify-deps verify-internal lint bench examples run-basic run-quickstart clean

GO ?= go
GOLANGCI_LINT ?= golangci-lint

all: lint test build ## 运行检查、测试与构建

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

tidy: ## 整理 go.mod
	$(GO) mod tidy

fmt: ## 格式化代码
	$(GO) fmt ./...

build: ## 编译全部包
	$(GO) build ./...

test: ## 运行单元测试
	$(GO) test -timeout=120s ./...

test-race: ## 运行竞态检测测试
	$(GO) test -race -timeout=120s ./...

test-race-pkg: ## 对指定包运行竞态测试，例如 make test-race-pkg PKG=./store/
	$(GO) test -race -timeout=60s -v $(PKG)

build-race-pkg: ## 对指定包构建竞态二进制，例如 make build-race-pkg PKG=./fetch/
	$(GO) build -race -x $(PKG)

verify-deps: ## 验证无第三方依赖
	@bad=$$($(GO) list -deps ./... | grep -E '\.(com|org|io|dev|net)/' | grep -v '^github.com/cysnb/go-sensitive-word' | grep -v '^vendor/' || true); \
	if [ -n "$$bad" ]; then echo "$$bad"; exit 1; fi; \
	echo "zero third-party deps OK"

verify-internal: ## 验证外部模块无法导入 internal 包
	@rm -rf .tmp-internal-check; mkdir -p .tmp-internal-check; \
	printf 'module example.com/check\n\ngo 1.21\n\nrequire github.com/cysnb/go-sensitive-word v0.0.0\n\nreplace github.com/cysnb/go-sensitive-word => %s\n' "$(CURDIR)" > .tmp-internal-check/go.mod; \
	printf 'package main\n\nimport _ "github.com/cysnb/go-sensitive-word/internal/normalize"\n\nfunc main() {}\n' > .tmp-internal-check/main.go; \
	if (cd .tmp-internal-check && $(GO) build ./... >/dev/null 2>&1); then \
		echo "internal import unexpectedly succeeded"; rm -rf .tmp-internal-check; exit 1; \
	fi; \
	rm -rf .tmp-internal-check; echo "internal visibility OK"

lint: ## 运行静态检查
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		$(GOLANGCI_LINT) run --timeout=5m ./...; \
	else \
		$(GO) vet ./...; \
	fi

bench: ## 运行基准测试
	$(GO) test -bench=. -benchmem ./...

examples: ## 编译全部示例
	$(GO) build ./examples/...

run-basic: ## 运行基础示例
	$(GO) run ./examples/basic

run-quickstart: ## 运行 README 快速开始示例
	$(GO) run ./examples/quickstart

clean: ## 清理构建产物
	$(GO) clean
	rm -f coverage.out coverage.html
