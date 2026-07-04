# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=1.0.0" \
    -o /spark-debug-mcp ./cmd/server/

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates kubectl \
    && addgroup -S spark && adduser -S spark -G spark

COPY --from=builder /spark-debug-mcp /usr/local/bin/spark-debug-mcp

USER spark

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD spark-debug-mcp version || exit 1

ENTRYPOINT ["spark-debug-mcp"]
CMD ["serve"]
