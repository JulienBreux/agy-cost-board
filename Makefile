.PHONY: all build build-web test cover cover-html lint vulncheck fmt tidy clean run demo

BINARY_NAME=agy-cost-board
CMD_DIR=./cmd/agy-cost-board
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(BUILD_DATE)

all: build

build-web:
	npm --prefix web ci
	npm --prefix web run build

build:
	go build -trimpath -v -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) $(CMD_DIR)

test:
	go test -v -race -coverprofile=coverage.out ./...

cover: test
	go tool cover -func=coverage.out

cover-html: test
	go tool cover -html=coverage.out -o coverage.html

lint:
	go tool golangci-lint run ./...

vulncheck:
	go tool govulncheck ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

clean:
	rm -rf bin/ coverage.out coverage.html web/dist

run:
	go run $(CMD_DIR)

demo:
	go run $(CMD_DIR) cost --demo

