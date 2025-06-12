VERSION := $$(jq -r .version version.json)
BRANCH := $$(git rev-parse --abbrev-ref HEAD)
HASH := $$(git rev-parse --short=7 HEAD)

LDFLAGS := "-X main.version=$(VERSION) -X main.branch=$(BRANCH) -X main.hash=$(HASH)"

OUTPUT := neko-engine

.PHONY:	all build dev clean

all: build

build:
	@go build -ldflags $(LDFLAGS) -o $(OUTPUT) app.go

dev:
	@go run -ldflags $(LDFLAGS) app.go serve -d
