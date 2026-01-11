# Quick Start - Long-Running Service Mode

This example demonstrates how to run the Alpha Detector as a long-running service.

## Prerequisites

- Go 1.21+ installed
- Or Docker installed

## Option 1: Run Locally

### Build
```bash
cd /home/runner/work/gp/gp
make build-service
```

### Run in All Mode (API + Scheduler)
```bash
export RUN_MODE=all
export API_PORT=8080
export LOG_FORMAT=pretty
export LOG_LEVEL=info

./bin/alpha-detector-service
```

Output:
```
{"level":"info","time":"2024-01-01T00:00:00Z","message":"Logger initialized","level":"info"}
{"level":"info","version":"2.0.0","build_time":"2024-01-01T00:00:00Z","message":"Starting Alpha Detector Service"}
{"level":"info","run_mode":"all","message":"Service run mode"}
{"level":"info","port":"8080","message":"Starting API server"}
{"level":"info","message":"Starting background scheduler"}
{"level":"info","message":"API server listening on :8080"}
```

### Test the API
```bash
# Health check
curl http://localhost:8080/health

# System status
curl http://localhost:8080/api/v1/status

# Trigger a scan job
curl -X POST http://localhost:8080/api/v1/jobs/scan \
  -H "Content-Type: application/json" \
  -d '{"stock_code":"600519","params":{"days":30}}'
```

## Option 2: Run with Docker

### Build Docker Image
```bash
docker build -f Dockerfile.service -t alpha-detector-service:latest .
```

### Run Container
```bash
docker run -d \
  --name alpha-service \
  -p 8080:8080 \
  -e RUN_MODE=all \
  -e LOG_LEVEL=info \
  -e ALPHA_KIMI_API_KEY=your-key-here \
  alpha-detector-service:latest

# View logs
docker logs -f alpha-service

# Stop
docker stop alpha-service
docker rm alpha-service
```

## Option 3: Different Run Modes

### API Only
```bash
export RUN_MODE=api
export API_PORT=8080
./bin/alpha-detector-service
```

### Scheduler Only
```bash
export RUN_MODE=scheduler
export SCAN_INTERVAL=1h
export PREDICT_INTERVAL=6h
./bin/alpha-detector-service
```

### Worker Only
```bash
export RUN_MODE=worker
export WORKER_POLL_INTERVAL=10s
./bin/alpha-detector-service
```

## Graceful Shutdown

Press Ctrl+C or send SIGTERM:
```bash
kill -TERM <pid>
```

The service will:
1. Stop accepting new requests
2. Complete in-flight requests (max 30s)
3. Stop background tasks
4. Exit cleanly

## Environment Variables Reference

| Variable | Description | Default |
|----------|-------------|---------|
| RUN_MODE | Run mode (all\|api\|scheduler\|worker) | all |
| API_PORT | API server port | 8080 |
| LOG_LEVEL | Log level (debug\|info\|warn\|error) | info |
| LOG_FORMAT | Log format (json\|pretty) | json |
| CONFIG_PATH | Configuration file path | configs/config.yaml |
| SCAN_INTERVAL | Scan interval (e.g., 1h, 30m) | 1h |
| PREDICT_INTERVAL | Predict interval | 6h |
| RUN_ON_STARTUP | Run scan on startup | true |

## Next Steps

- See [docs/SERVICE_DEPLOYMENT.md](SERVICE_DEPLOYMENT.md) for complete deployment guide
- See [docs/DATABRICKS_ARM64_DEPLOYMENT.md](DATABRICKS_ARM64_DEPLOYMENT.md) for Databricks deployment
- Configure external services (Redis, PostgreSQL, Kafka) via environment variables
