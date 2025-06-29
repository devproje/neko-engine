VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || echo "v0.0.0")
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION := $(shell go version | cut -d' ' -f3)

BINARY_NAME := neko-engine
MAIN_PATH := ./cmd

DIST_DIR := dist
BUILD_DIR := build

TC_BUILD_NUMBER ?= 0
TC_BUILD_BRANCH ?= $(BRANCH)

RELEASE_CHANNEL := $(shell if [ "$(BRANCH)" = "master" ]; then echo "stable"; elif [ "$(BRANCH)" = "nightly" ]; then echo "nightly"; else echo "dev"; fi)

VERSION_SUFFIX := $(shell if [ "$(RELEASE_CHANNEL)" = "stable" ]; then echo ""; elif [ "$(RELEASE_CHANNEL)" = "nightly" ]; then echo "-nightly.$$(date +%Y%m%d)"; else echo "-dev.$(HASH)"; fi)

FULL_VERSION := $(VERSION)$(VERSION_SUFFIX)
ifneq ($(TC_BUILD_NUMBER),0)
	FULL_VERSION := $(VERSION)$(VERSION_SUFFIX).$(TC_BUILD_NUMBER)
endif

LDFLAGS := -X main.version=$(FULL_VERSION) \
           -X main.branch=$(BRANCH) \
           -X main.hash=$(HASH) \
           -X main.buildTime=$(BUILD_TIME) \
           -X main.goVersion=$(GO_VERSION) \
           -X main.buildNumber=$(TC_BUILD_NUMBER) \
           -X main.channel=$(RELEASE_CHANNEL)

BUILD_FLAGS := -ldflags "$(LDFLAGS)" -trimpath
RELEASE_FLAGS := -ldflags "$(LDFLAGS) -s -w" -trimpath

LINUX_PLATFORMS := linux/amd64 linux/arm64
DARWIN_PLATFORMS := darwin/amd64 darwin/arm64
WINDOWS_PLATFORMS := windows/amd64 windows/arm64
ALL_PLATFORMS := $(LINUX_PLATFORMS) $(DARWIN_PLATFORMS) $(WINDOWS_PLATFORMS)

.PHONY: all build clean test lint check linux-build darwin-build windows-build cross-compile artifacts release version ci-prepare ci-build ci-test ci-artifacts ci help

all: build

build:
	@echo "Building $(BINARY_NAME) (dev)..."
	@go build $(BUILD_FLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(DIST_DIR) $(BUILD_DIR)
	@go clean -cache -modcache -testcache

test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

lint:
	@echo "Running linter..."
	@go vet ./...
	@go fmt ./...
	@go mod tidy

check: lint test
	@echo "All checks passed"

linux-build:
	@echo "Building for Linux platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(LINUX_PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		echo "Building for $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH CGO_ENABLED=0 go build $(RELEASE_FLAGS) \
			-o $(DIST_DIR)/$(BINARY_NAME)-$$GOOS-$$GOARCH $(MAIN_PATH); \
	done

darwin-build:
	@echo "Building for macOS platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(DARWIN_PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		echo "Building for $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH CGO_ENABLED=0 go build $(RELEASE_FLAGS) \
			-o $(DIST_DIR)/$(BINARY_NAME)-$$GOOS-$$GOARCH $(MAIN_PATH); \
	done

windows-build:
	@echo "Building for Windows platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(WINDOWS_PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		echo "Building for $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH CGO_ENABLED=0 go build $(RELEASE_FLAGS) \
			-o $(DIST_DIR)/$(BINARY_NAME)-$$GOOS-$$GOARCH.exe $(MAIN_PATH); \
	done

cross-compile: linux-build darwin-build windows-build
	@echo "Cross-compilation completed for all platforms"

artifacts: cross-compile
	@echo "Binaries ready in $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/

release: check artifacts
	@echo "Release $(FULL_VERSION) completed successfully"
	@ls -la $(DIST_DIR)/

version:
	@echo "Version: $(FULL_VERSION)"
	@echo "Channel: $(RELEASE_CHANNEL)"
	@echo "Branch: $(BRANCH)"
	@echo "Hash: $(HASH)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Build Number: $(TC_BUILD_NUMBER)"

ci-prepare:
	@echo "##teamcity[buildNumber '$(FULL_VERSION)']"
	@echo "Preparing CI environment..."
	@go version
	@go env

ci-build: ci-prepare cross-compile
	@echo "CI build completed for all platforms"

ci-test: ci-prepare test
	@echo "##teamcity[publishArtifacts 'coverage.out']"
	@echo "CI tests completed"

ci-artifacts: ci-prepare artifacts
	@echo "##teamcity[publishArtifacts '$(DIST_DIR)/*']"
	@echo "CI binaries published"

ci: ci-prepare check ci-build ci-test ci-artifacts
	@echo "##teamcity[buildStatus text='Multi-platform build $(FULL_VERSION) completed successfully']"

help:
	@echo "Neko Engine Build System"
	@echo ""
	@echo "Development targets:"
	@echo "  build        - Build binary for development"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests with coverage"
	@echo "  lint         - Run linter and formatter"
	@echo "  check        - Run all quality checks"
	@echo ""
	@echo "Release targets:"
	@echo "  linux-build  - Build for Linux platforms"
	@echo "  darwin-build - Build for macOS platforms"
	@echo "  windows-build- Build for Windows platforms"
	@echo "  cross-compile- Build for all platforms"
	@echo "  artifacts    - Prepare release binaries"
	@echo "  release      - Full release build"
	@echo ""
	@echo "CI targets (TeamCity):"
	@echo "  ci-prepare   - Prepare CI environment"
	@echo "  ci-build     - CI build step (all platforms)"
	@echo "  ci-test      - CI test step"
	@echo "  ci-artifacts - CI artifact publishing"
	@echo "  ci           - Full CI pipeline (all platforms)"
	@echo ""
	@echo "Utility:"
	@echo "  version      - Show version information"
	@echo "  help         - Show this help"
