.PHONY: build docker lint fmt test run clean version

VERSION ?= 1.0.0
BINARY  := bin/spark-debug-mcp
IMAGE   := spark-debug-mcp:$(VERSION)

build:
	@mkdir -p bin
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BINARY) ./cmd/server/

docker:
	docker build -t $(IMAGE) .

lint:
	@which golangci-lint > /dev/null || (echo "install golangci-lint: https://golangci-lint.run/welcome/install/" && exit 1)
	golangci-lint run ./...

fmt:
	go fmt ./...
	gofmt -s -w .

test:
	go test -race ./...
	go test -race -coverprofile=coverage.out ./internal/... ./pkg/...
	go tool cover -func=coverage.out | tail -1

run: build
	./$(BINARY) serve

clean:
	rm -rf bin/ coverage.out

version:
	@echo $(VERSION)
