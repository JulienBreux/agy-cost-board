.PHONY: all build test cover lint clean run demo

BINARY_NAME=agy-ge-board
CMD_DIR=./cmd/agy-ge-board

all: build

build:
	go build -v -o bin/$(BINARY_NAME) $(CMD_DIR)

test:
	go test -v -race -coverprofile=coverage.out ./...

cover: test
	go tool cover -func=coverage.out

cover-html: test
	go tool cover -html=coverage.out -o coverage.html

lint:
	go vet ./...

clean:
	rm -rf bin/ coverage.out coverage.html web/dist

run:
	go run $(CMD_DIR)

demo:
	go run $(CMD_DIR) cost --demo
