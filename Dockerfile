# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always 2>/dev/null || echo v1.0.0) -X main.BuildTime=$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
    -o /app/alpha-detector \
    ./cmd/alpha

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary and configs
COPY --from=builder /app/alpha-detector .
COPY configs/ ./configs/

# Create non-root user
RUN adduser -D -u 1000 appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8888 9090

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:9090/health || exit 1

ENTRYPOINT ["./alpha-detector"]
CMD ["scan", "--config", "configs/config.yaml"]
