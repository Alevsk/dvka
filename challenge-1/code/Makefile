# ---- config ----
GO        := go
PKGS      := ./...
MAIN_PKG  := ./cmd/app
BINARY    := lab1
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_TAG    := $(shell git describe --tags --exact-match 2>/dev/null || echo "")
VERSION    := $(if $(GIT_TAG),$(GIT_TAG),$(GIT_COMMIT))
LDFLAGS   := -s -w -X 'main.version=$(VERSION)' -X 'main.buildTime=$(BUILD_TIME)' -X 'main.commit=$(GIT_COMMIT)'

# Docker
IMAGE ?= alevsk/dvka
TAG   ?= lab-1-$(VERSION)
PLATFORMS ?= linux/amd64

# ---- phony targets ----
.PHONY: all build run clean fmt fmt-check imports vet lint staticcheck test cover tidy tidy-check docker-build docker-run

all: build

build:
	@echo "Building $(BINARY) -> ./cmd/app/$(BINARY)"
	@CGO_ENABLED=0 $(GO) build -trimpath -tags=kqueue -ldflags "$(LDFLAGS)" -o ./cmd/app/$(BINARY) $(MAIN_PKG)

run: build
	@./cmd/app/$(BINARY)

clean:
	@rm -f ./cmd/app/$(BINARY)

# Write format using go fmt (module-aware)
fmt:
	@$(GO) fmt $(PKGS)

# CI-friendly: fail if anything needs formatting or import fixing
fmt-check:
	@OUT=$$(gofmt -s -l .); if [ -n "$$OUT" ]; then echo "gofmt needed on:"; echo "$$OUT"; exit 1; fi

# Optional: deterministic imports (pick one tool; comment out the other)
imports:
	@which goimports >/dev/null 2>&1 && goimports -w . || true
#	@which gci >/dev/null 2>&1 && gci write . || true

vet:
	@$(GO) vet $(PKGS)

# Use one of these (golangci-lint recommended)
lint:
	@which golangci-lint >/dev/null 2>&1 && golangci-lint run || { echo "golangci-lint not installed"; exit 1; }

staticcheck:
	@which staticcheck >/dev/null 2>&1 && staticcheck $(PKGS) || { echo "staticcheck not installed"; exit 1; }

test:
	@$(GO) test -race -count=1 $(PKGS)

cover:
	@$(GO) test -race -coverprofile=coverage.out $(PKGS)
	@$(GO) tool cover -func=coverage.out | tail -n1

tidy:
	@$(GO) mod tidy

tidy-check:
	@TMP=$$(mktemp -d); cp go.mod go.sum $$TMP/; $(GO) mod tidy; \
		diff -u $$TMP/go.mod go.mod && diff -u $$TMP/go.sum go.sum; \
		rm -rf $$TMP

docker-build:
	@docker buildx build \
		--platform $(PLATFORMS) \
		--build-arg build_version=$(VERSION) \
		--build-arg build_time='$(BUILD_TIME)' \
		-t $(IMAGE):$(TAG) \
		--output=type=docker \
		.

docker-run:
	@docker run --rm -it -p 8080:8080 --name dvka-lab-1 $(IMAGE):$(TAG)
