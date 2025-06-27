VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION := $(shell go version | cut -d' ' -f3)

BINARY_NAME := neko-engine
MAIN_PATH := ./cmd

LDFLAGS := -X main.version=$(FULL_VERSION) \
           -X main.branch=$(BRANCH) \
           -X main.hash=$(HASH) \
           -X main.buildTime=$(BUILD_TIME) \
           -X main.goVersion=$(GO_VERSION) \
           -X main.channel=$(RELEASE_CHANNEL)

BUILD_FLAGS := -ldflags "$(LDFLAGS)"

.PHONY: all build clean install test fmt vet mod-tidy help

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(BUILD_FLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@go clean

install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(BUILD_FLAGS) $(MAIN_PATH)

test:
	@echo "Running tests..."
	@go test -v ./...

fmt:
	@echo "Formatting code..."
	@go fmt ./...

vet:
	@echo "Vetting code..."
	@go vet ./...

mod-tidy:
	@echo "Tidying modules..."
	@go mod tidy

help:
	@echo "Available targets:"
	@echo "  build     - Build the binary"
	@echo "  clean     - Clean build artifacts"
	@echo "  install   - Install the binary"
	@echo "  test      - Run tests"
	@echo "  fmt       - Format code"
	@echo "  vet       - Vet code"
	@echo "  mod-tidy  - Tidy modules"
	@echo "  tag-patch    - Create patch version tag (v1.0.X) [master only]"
	@echo "  tag-minor    - Create minor version tag (v1.X.0) [master only]"
	@echo "  tag-major    - Create major version tag (vX.0.0) [master only]"
	@echo "  release      - Build release for current channel"
	@echo "  version-info - Show detailed version information"
	@echo "  help         - Show this help"
	@echo ""
	@echo "Release Channels:"
	@echo "  master/main  -> stable (v1.0.0)"
	@echo "  nightly      -> nightly (v1.0.0-nightly.20240101)"
	@echo "  dev/develop  -> dev (v1.0.0-dev.abc1234)"

# Git tag-based version control and release channels
CURRENT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
CURRENT_VERSION := $(shell echo $(CURRENT_TAG) | sed 's/^v//')

# Release channel configuration
RELEASE_CHANNEL := $(shell \
	case "$(BRANCH)" in \
		master|main) echo "stable" ;; \
		nightly) echo "nightly" ;; \
		dev|develop) echo "dev" ;; \
		*) echo "dev" ;; \
	esac)

# Channel-specific version suffix
VERSION_SUFFIX := $(shell \
	case "$(RELEASE_CHANNEL)" in \
		stable) echo "" ;; \
		nightly) echo "-nightly.$(shell date +%Y%m%d)" ;; \
		dev) echo "-dev.$(HASH)" ;; \
		*) echo "-dev.$(HASH)" ;; \
	esac)

FULL_VERSION := $(VERSION)$(VERSION_SUFFIX)

tag-patch:
	@echo "Current version: $(CURRENT_TAG)"
	@if [ "$(BRANCH)" != "master" ] && [ "$(BRANCH)" != "main" ]; then \
		echo "Error: Tags can only be created from master/main branch. Current branch: $(BRANCH)"; \
		exit 1; \
	fi
	@NEW_VERSION=$$(echo $(CURRENT_VERSION) | awk -F. '{$$3++; print $$1"."$$2"."$$3}'); \
	echo "Creating patch tag: v$$NEW_VERSION"; \
	git tag -a "v$$NEW_VERSION" -m "Release v$$NEW_VERSION"; \
	echo "Tag v$$NEW_VERSION created"

tag-minor:
	@echo "Current version: $(CURRENT_TAG)"
	@if [ "$(BRANCH)" != "master" ] && [ "$(BRANCH)" != "main" ]; then \
		echo "Error: Tags can only be created from master/main branch. Current branch: $(BRANCH)"; \
		exit 1; \
	fi
	@NEW_VERSION=$$(echo $(CURRENT_VERSION) | awk -F. '{$$2++; $$3=0; print $$1"."$$2"."$$3}'); \
	echo "Creating minor tag: v$$NEW_VERSION"; \
	git tag -a "v$$NEW_VERSION" -m "Release v$$NEW_VERSION"; \
	echo "Tag v$$NEW_VERSION created"

tag-major:
	@echo "Current version: $(CURRENT_TAG)"
	@if [ "$(BRANCH)" != "master" ] && [ "$(BRANCH)" != "main" ]; then \
		echo "Error: Tags can only be created from master/main branch. Current branch: $(BRANCH)"; \
		exit 1; \
	fi
	@NEW_VERSION=$$(echo $(CURRENT_VERSION) | awk -F. '{$$1++; $$2=0; $$3=0; print $$1"."$$2"."$$3}'); \
	echo "Creating major tag: v$$NEW_VERSION"; \
	git tag -a "v$$NEW_VERSION" -m "Release v$$NEW_VERSION"; \
	echo "Tag v$$NEW_VERSION created"

release: build
	@echo "Building release for channel: $(RELEASE_CHANNEL)"
	@echo "Full version: $(FULL_VERSION)"
	@echo "Branch: $(BRANCH)"
	@echo "Release $(FULL_VERSION) built successfully"

version-info:
	@echo "Version: $(FULL_VERSION)"
	@echo "Channel: $(RELEASE_CHANNEL)"
	@echo "Branch: $(BRANCH)"
	@echo "Hash: $(HASH)"
	@echo "Build Time: $(BUILD_TIME)"

.PHONY: tag-patch tag-minor tag-major release version-info
