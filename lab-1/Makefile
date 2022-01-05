PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)
# Sets the build version based on the output of the following command, if we are building for a tag, that's the build else it uses the current git branch as the build
BUILD_VERSION:=$(shell git describe --exact-match --tags $(git log -n1 --pretty='%h') 2>/dev/null || git rev-parse --abbrev-ref HEAD 2>/dev/null)
BUILD_TIME:=$(shell date 2>/dev/null)
TAG ?= "alevsk/dvka:lab-1-$(BUILD_VERSION)"

default: lab1

.PHONY: lab1
lab1:
	@echo "Building Lab 1 binary to './cmd/app/lab1'"
	@(GO111MODULE=on CGO_ENABLED=0 go build -trimpath --tags=kqueue --ldflags "-s -w" -o ./cmd/app/lab1 ./cmd/app)

verifiers: fmt

fmt:
	@echo "Running $@ check"
	@GO111MODULE=on gofmt -d cmd/

docker:
	@docker buildx build --output=type=docker --platform linux/amd64 -t $(TAG) --build-arg build_version=$(BUILD_VERSION) --build-arg build_time='$(BUILD_TIME)' .
