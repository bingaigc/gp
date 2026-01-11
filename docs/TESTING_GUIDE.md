# Testing the Alpha Detector Service

This guide provides comprehensive testing instructions for the Alpha Detector Service in different deployment scenarios.

## Table of Contents

1. [Local Testing (Binary)](#local-testing-binary)
2. [Docker Testing](#docker-testing)
3. [Multi-arch Build Testing](#multi-arch-build-testing)
4. [API Testing](#api-testing)
5. [Service Mode Testing](#service-mode-testing)
6. [Databricks Testing](#databricks-testing)

## Prerequisites

```bash
# Install dependencies
go mod download

# Build tools
make --version
docker --version
curl --version
```

## Local Testing (Binary)

### 1. Build the Binaries

```bash
# Build both CLI and Service
make build-all

# Or build individually
make build-cli      # CLI binary
make build-service  # Service binary
```

### 2. Test CLI Mode

```bash
# Version check
./bin/alpha-detector version

# Health check
./bin/alpha-detector health

# Demo scan (no API key required)
./bin/alpha-detector scan

# With API key
export ALPHA_KIMI_API_KEY="your-api-key"
./bin/alpha-detector scan-real
```

### 3. Test Service Mode

**Terminal 1: Start service**
```bash
export RUN_MODE=all
export API_PORT=8080
export LOG_LEVEL=info
export LOG_FORMAT=pretty

./bin/alpha-detector-service
```

**Terminal 2: Test API**
```bash
# Health check
curl http://localhost:8080/health

# System status
curl http://localhost:8080/api/v1/status

# Version
curl http://localhost:8080/api/v1/version

# Trigger scan
curl -X POST http://localhost:8080/api/v1/jobs/scan \
  -H "Content-Type: application/json" \
  -d '{"stock_code":"600519"}'
```

## Docker Testing

### 1. Build Docker Image

```bash
# Build service image
docker build -f Dockerfile.service -t alpha-detector-service:test .

# Verify image
docker images | grep alpha-detector
```

### 2. Run Container Locally

```bash
# Run with all modes
docker run -d \
  --name alpha-service-test \
  -p 8080:8080 \
  -e RUN_MODE=all \
  -e LOG_LEVEL=info \
  -e ALPHA_KIMI_API_KEY="your-api-key" \
  alpha-detector-service:test

# Check logs
docker logs -f alpha-service-test

# Test API
curl http://localhost:8080/health
```

### 3. Test Different Run Modes

**API-only mode:**
```bash
docker run -d -p 8080:8080 \
  -e RUN_MODE=api \
  alpha-detector-service:test
```

**Scheduler-only mode:**
```bash
docker run -d \
  -e RUN_MODE=scheduler \
  -e SCAN_INTERVAL=5m \
  alpha-detector-service:test
```

**Worker-only mode:**
```bash
docker run -d \
  -e RUN_MODE=worker \
  -e WORKER_CONCURRENCY=5 \
  alpha-detector-service:test
```

### 4. Cleanup

```bash
docker stop alpha-service-test
docker rm alpha-service-test
docker rmi alpha-detector-service:test
```

## Multi-arch Build Testing

### 1. Setup Buildx

```bash
# Create builder
docker buildx create --name multiarch --use

# Verify
docker buildx inspect --bootstrap
```

### 2. Build Multi-arch Image

```bash
# Build for AMD64 and ARM64
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f Dockerfile.service \
  -t your-registry/alpha-detector-service:test \
  --push .
```

### 3. Test ARM64 (on ARM64 machine)

```bash
# Pull and run ARM64 image
docker pull --platform linux/arm64 your-registry/alpha-detector-service:test
docker run -d -p 8080:8080 your-registry/alpha-detector-service:test

# Verify architecture
docker inspect your-registry/alpha-detector-service:test | grep Architecture
```

### 4. Test AMD64 (on x86_64 machine)

```bash
# Pull and run AMD64 image
docker pull --platform linux/amd64 your-registry/alpha-detector-service:test
docker run -d -p 8080:8080 your-registry/alpha-detector-service:test
```

## API Testing

### Complete API Test Suite

```bash
#!/bin/bash
BASE_URL="http://localhost:8080"

echo "=== Testing Alpha Detector API ==="

# 1. Health check
echo -e "\n1. Health Check:"
curl -s $BASE_URL/health | jq .

# 2. Readiness check
echo -e "\n2. Readiness Check:"
curl -s $BASE_URL/ready | jq .

# 3. Version info
echo -e "\n3. Version:"
curl -s $BASE_URL/api/v1/version | jq .

# 4. System status
echo -e "\n4. Status:"
curl -s $BASE_URL/api/v1/status | jq .

# 5. Trigger scan
echo -e "\n5. Trigger Scan:"
curl -s -X POST $BASE_URL/api/v1/jobs/scan \
  -H "Content-Type: application/json" \
  -d '{"stock_code":"600519","count":10}' | jq .

# 6. Trigger prediction
echo -e "\n6. Trigger Prediction:"
curl -s -X POST $BASE_URL/api/v1/jobs/predict \
  -H "Content-Type: application/json" \
  -d '{"stock_code":"600519","indicator":"MA5"}' | jq .

# 7. Trigger backtest
echo -e "\n7. Trigger Backtest:"
curl -s -X POST $BASE_URL/api/v1/jobs/backtest \
  -H "Content-Type: application/json" \
  -d '{"stock_code":"600519","days":365}' | jq .

# 8. Query signals
echo -e "\n8. Query Signals:"
curl -s "$BASE_URL/api/v1/signals?limit=5" | jq .

# 9. Get metrics
echo -e "\n9. Metrics:"
curl -s $BASE_URL/api/v1/metrics | jq .

echo -e "\n=== API Tests Complete ==="
```

Save as `test_api.sh` and run:
```bash
chmod +x test_api.sh
./test_api.sh
```

### Load Testing

```bash
# Install apache bench
sudo apt-get install apache2-utils  # Ubuntu/Debian
brew install httpd                   # macOS

# Run load test (1000 requests, 10 concurrent)
ab -n 1000 -c 10 http://localhost:8080/health

# Run load test with JSON payload
ab -n 100 -c 5 -p payload.json -T application/json \
   http://localhost:8080/api/v1/jobs/scan
```

## Service Mode Testing

### Test Graceful Shutdown

```bash
# Start service
./bin/alpha-detector-service &
SERVICE_PID=$!

# Wait a bit
sleep 5

# Send SIGTERM (graceful shutdown)
kill -TERM $SERVICE_PID

# Check logs for graceful shutdown message
# Should see: "Shutting down gracefully..."
```

### Test Signal Handling

```bash
# Test SIGINT (Ctrl+C)
./bin/alpha-detector-service
# Press Ctrl+C and verify graceful shutdown

# Test SIGTERM
./bin/alpha-detector-service &
kill -TERM $!
```

### Test All Run Modes

```bash
# Test each mode for 30 seconds
for mode in all api scheduler worker; do
  echo "Testing RUN_MODE=$mode"
  RUN_MODE=$mode ./bin/alpha-detector-service &
  PID=$!
  sleep 30
  kill -TERM $PID
  wait $PID
  echo "Mode $mode completed successfully"
done
```

## Databricks Testing

### 1. Build and Push Image

```bash
# Build multi-arch image
docker buildx build \
  --platform linux/arm64 \
  -f Dockerfile.service \
  -t your-registry/alpha-detector-service:databricks-test \
  --push .
```

### 2. Create Databricks Secrets

```bash
databricks secrets create-scope alpha-detector
databricks secrets put-secret --scope alpha-detector --key kimi-api-key --string-value "your-key"
```

### 3. Upload Monitor Script

```bash
databricks fs cp deploy/databricks/keep_alive.py dbfs:/databricks/scripts/keep_alive.py
```

### 4. Create Test Job

Edit `databricks_job_template.json` and create job:
```bash
databricks jobs create --json-file deploy/databricks/databricks_job_template.json
```

### 5. Run and Monitor

```bash
# Run job
databricks jobs run-now --job-id YOUR_JOB_ID

# Get run ID from output, then check status
databricks runs get --run-id YOUR_RUN_ID

# View logs
databricks runs get-output --run-id YOUR_RUN_ID
```

### 6. Test from Notebook

Create a Databricks notebook:

```python
import requests
import json

# Test health endpoint
response = requests.get("http://localhost:8080/health")
print("Health:", response.json())

# Test status
response = requests.get("http://localhost:8080/api/v1/status")
print("Status:", json.dumps(response.json(), indent=2))

# Trigger scan
response = requests.post(
    "http://localhost:8080/api/v1/jobs/scan",
    json={"stock_code": "600519", "count": 10}
)
print("Scan result:", json.dumps(response.json(), indent=2))
```

## Performance Testing

### 1. Memory Usage

```bash
# Monitor memory while running
docker stats alpha-service-test

# Or for local binary
ps aux | grep alpha-detector-service
```

### 2. CPU Usage

```bash
# Top command
top -p $(pgrep alpha-detector)

# Docker stats
docker stats alpha-service-test
```

### 3. Response Time Benchmarking

```bash
# Install hey (HTTP load generator)
go install github.com/rakyll/hey@latest

# Benchmark health endpoint
hey -n 10000 -c 100 http://localhost:8080/health

# Benchmark scan endpoint
hey -n 1000 -c 50 -m POST \
  -H "Content-Type: application/json" \
  -d '{"stock_code":"600519"}' \
  http://localhost:8080/api/v1/jobs/scan
```

## Integration Testing

### With Redis

```bash
# Start Redis
docker run -d --name redis-test -p 6379:6379 redis:alpine

# Run service with Redis
export REDIS_URL="redis://localhost:6379"
./bin/alpha-detector-service

# Verify connection in logs
```

### With PostgreSQL

```bash
# Start PostgreSQL
docker run -d --name postgres-test \
  -e POSTGRES_PASSWORD=testpass \
  -e POSTGRES_DB=alpha_detector \
  -p 5432:5432 postgres:alpine

# Run service with PostgreSQL
export POSTGRES_URL="postgresql://postgres:testpass@localhost:5432/alpha_detector"
./bin/alpha-detector-service
```

### With Kafka

```bash
# Start Kafka (with docker-compose)
docker-compose -f deploy/docker-compose.yml up -d kafka zookeeper

# Run service with Kafka
export KAFKA_BROKERS="localhost:9092"
./bin/alpha-detector-service
```

## Automated Testing

### Run All Tests

```bash
# Unit tests
make test

# Integration tests (if available)
make test-integration

# Build tests
make build-all

# Docker build test
make docker-build

# Full test suite
make test-all
```

### CI/CD Testing

GitHub Actions will automatically:
1. Build binaries
2. Run unit tests
3. Build Docker images
4. Run integration tests

Check `.github/workflows/ci.yml` for details.

## Troubleshooting Tests

### Service won't start

```bash
# Check port availability
lsof -i :8080

# Check configuration
./bin/alpha-detector-service --help

# Run with debug logging
LOG_LEVEL=debug ./bin/alpha-detector-service
```

### API returns errors

```bash
# Check logs
docker logs alpha-service-test

# Verbose curl
curl -v http://localhost:8080/health

# Check service status
curl http://localhost:8080/api/v1/status
```

### Docker build fails

```bash
# Clean build
docker system prune -a
docker build --no-cache -f Dockerfile.service -t alpha-detector-service:test .

# Check Docker daemon
docker info
```

## Test Checklist

Before production deployment, verify:

- [ ] CLI version command works
- [ ] CLI scan command works
- [ ] Service starts successfully
- [ ] Health endpoint responds
- [ ] All API endpoints accessible
- [ ] Service handles SIGTERM gracefully
- [ ] Docker image builds successfully
- [ ] Multi-arch images work on target platforms
- [ ] Databricks job runs successfully
- [ ] API can trigger all job types
- [ ] Logs are structured and readable
- [ ] Performance is acceptable under load
- [ ] Memory usage is stable
- [ ] No resource leaks after extended operation

## Summary

You've tested:
- ✅ Local binary execution
- ✅ Docker container operation
- ✅ Multi-architecture builds
- ✅ All API endpoints
- ✅ Different service modes
- ✅ Databricks deployment
- ✅ Performance and load
- ✅ Integration with external services

The system is ready for production deployment!
