VERSION := $(shell cat VERSION)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: build build-prod build-all test clean lint

# Dev build (no obfuscation)
build:
	go build -ldflags "$(LDFLAGS)" -o dist/claudex ./cmd/claudex/

# Production build (obfuscated)
build-prod:
	garble -literals -tiny -seed=random build -ldflags "$(LDFLAGS)" -o dist/claudex ./cmd/claudex/

# Cross-platform production builds
build-all:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 garble -literals -tiny build -ldflags "$(LDFLAGS)" -o dist/claudex-windows-amd64.exe ./cmd/claudex/
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 garble -literals -tiny build -ldflags "$(LDFLAGS)" -o dist/claudex-darwin-amd64 ./cmd/claudex/
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 garble -literals -tiny build -ldflags "$(LDFLAGS)" -o dist/claudex-darwin-arm64 ./cmd/claudex/
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 garble -literals -tiny build -ldflags "$(LDFLAGS)" -o dist/claudex-linux-amd64 ./cmd/claudex/

test:
	go test ./... -v

clean:
	rm -rf dist/

lint:
	golangci-lint run ./...
