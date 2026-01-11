# Long-Running Service Mode - Deployment Guide

## Overview

The Alpha Detector system has been transformed into a **long-running service** with multiple run modes:

- **API Mode**: REST API server for health checks and job triggering
- **Scheduler Mode**: Background task scheduler for periodic scans
- **Worker Mode**: Job processing worker
- **All Mode** (default): Runs both API and Scheduler

## Architecture

```
┌─────────────────────────────────────────┐
│     Alpha Detector Service              │
│                                          │
│  ┌────────────┐      ┌───────────────┐ │
│  │ API Server │      │   Scheduler   │ │
│  │  (Gin)     │      │  (Cron-like)  │ │
│  │            │      │               │ │
│  │ :8080      │      │ Periodic      │ │
│  │            │      │ Tasks         │ │
│  └────────────┘      └───────────────┘ │
│         │                     │         │
│         └─────────┬───────────┘         │
│                   │                     │
│         ┌─────────▼──────────┐         │
│         │  Shared Services   │         │
│         │  (Config, Logger)  │         │
│         └────────────────────┘         │
└─────────────────────────────────────────┘
```

## Run Modes

### 1. All Mode (Default)
Runs both API server and scheduler in the same process.

```bash
export RUN_MODE=all
./bin/alpha-detector-service
```

### 2. API Mode
Runs only the REST API server.

```bash
export RUN_MODE=api
export API_PORT=8080
./bin/alpha-detector-service
```

### 3. Scheduler Mode
Runs only the background scheduler.

```bash
export RUN_MODE=scheduler
export SCAN_INTERVAL=1h
export PREDICT_INTERVAL=6h
./bin/alpha-detector-service
```

### 4. Worker Mode
Runs only the worker for job processing.

```bash
export RUN_MODE=worker
export WORKER_POLL_INTERVAL=10s
./bin/alpha-detector-service
```

## Environment Variables

### Core Settings
- `RUN_MODE`: Service run mode (all|api|scheduler|worker) [default: all]
- `CONFIG_PATH`: Path to configuration file [default: configs/config.yaml]
- `LOG_LEVEL`: Logging level (debug|info|warn|error) [default: info]
- `LOG_FORMAT`: Log format (json|pretty) [default: json]

### API Server Settings
- `API_PORT`: API server port [default: 8080]
- `GIN_MODE`: Gin mode (debug|release) [default: release]

### Scheduler Settings
- `SCAN_INTERVAL`: Scan task interval (e.g., 1h, 30m) [default: 1h]
- `PREDICT_INTERVAL`: Predict task interval [default: 6h]
- `RUN_ON_STARTUP`: Run initial scan on startup [default: true]

### Worker Settings
- `WORKER_POLL_INTERVAL`: Job poll interval [default: 10s]
- `WORKER_CONCURRENCY`: Number of concurrent workers [default: 3]

### External Services
- `REDIS_URL`: Redis connection URL [default: redis://localhost:6379]
- `POSTGRES_URL`: PostgreSQL connection URL
- `KAFKA_BROKERS`: Kafka broker addresses
- `ALPHA_KIMI_API_KEY`: Kimi AI API key

## API Endpoints

### Health & Status
```bash
# Health check
GET /health
Response: {"status":"healthy","timestamp":"2024-01-01T00:00:00Z","version":"2.0.0"}

# Readiness check
GET /ready
Response: {"status":"ready","checks":{"config":"ok","db":"ok"}}

# System status
GET /api/v1/status
Response: {"status":"running","uptime":"1h30m","goroutines":25,"memory_mb":128}

# Version info
GET /api/v1/version
Response: {"version":"2.0.0","build_time":"2024-01-01","go_version":"go1.21.0"}
```

### Job Triggers
```bash
# Trigger scan job
POST /api/v1/jobs/scan
Body: {"stock_code":"600519","params":{"days":30}}
Response: {"job_id":"20240101-120000-abc123","status":"queued"}

# Trigger predict job
POST /api/v1/jobs/predict
Body: {"stock_code":"600519","params":{"indicator":"MA5"}}

# Trigger backtest job
POST /api/v1/jobs/backtest
Body: {"stock_code":"600519","params":{"days":365,"capital":1000000}}
```

### Queries
```bash
# Get signals
GET /api/v1/signals?limit=10&offset=0

# Get metrics
GET /api/v1/metrics
```

## Building

### Build CLI (Legacy)
```bash
make build
# Output: bin/alpha-detector
```

### Build Service
```bash
go build -o bin/alpha-detector-service ./cmd/service
```

### Build Docker Image
```bash
docker build -t alpha-detector-service:latest .
```

## Local Deployment

### Using Binary
```bash
# Start in all mode
export RUN_MODE=all
export API_PORT=8080
export LOG_FORMAT=pretty
./bin/alpha-detector-service
```

### Using Docker
```bash
docker run -d \
  --name alpha-service \
  -p 8080:8080 \
  -e RUN_MODE=all \
  -e LOG_LEVEL=info \
  -e ALPHA_KIMI_API_KEY=your-key \
  alpha-detector-service:latest
```

## Databricks Deployment

### 1. Single Container Mode

Create a Databricks Job with the following configuration:

```python
# Databricks Job Configuration
job_config = {
    "name": "alpha-detector-service",
    "new_cluster": {
        "spark_version": "13.3.x-scala2.12",
        "node_type_id": "m6g.xlarge",  # ARM64 instance
        "num_workers": 0,  # Single node
        "custom_tags": {
            "project": "alpha-detector"
        }
    },
    "docker_image": {
        "url": "your-registry/alpha-detector-service:latest",
        "basic_auth": {
            "username": "{{secrets/docker/username}}",
            "password": "{{secrets/docker/password}}"
        }
    },
    "spark_env_vars": {
        "RUN_MODE": "all",
        "API_PORT": "8080",
        "LOG_LEVEL": "info",
        "ALPHA_KIMI_API_KEY": "{{secrets/alpha-detector/kimi-api-key}}",
        "REDIS_URL": "{{secrets/alpha-detector/redis-url}}",
        "POSTGRES_URL": "{{secrets/alpha-detector/postgres-url}}"
    }
}
```

### 2. Init Script Deployment

Create an init script at `dbfs:/databricks/scripts/alpha-detector-init.sh`:

```bash
#!/bin/bash
# Install Go
wget https://go.dev/dl/go1.21.0.linux-arm64.tar.gz
tar -C /usr/local -xzf go1.21.0.linux-arm64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Clone and build
cd /tmp
git clone https://github.com/bingaigc/gp.git
cd gp
export GOARCH=arm64 GOOS=linux
make build-service

# Copy binary to dbfs
cp bin/alpha-detector-service /dbfs/alpha-detector/bin/

# Start service in background
nohup /dbfs/alpha-detector/bin/alpha-detector-service > /tmp/alpha.log 2>&1 &
```

### 3. Notebook Deployment

```python
# Databricks Notebook
# Cell 1: Install and start service
%sh
export RUN_MODE=api
export API_PORT=8080
/dbfs/alpha-detector/bin/alpha-detector-service &

# Cell 2: Wait for service to start
import time
import requests

time.sleep(5)
response = requests.get("http://localhost:8080/health")
print(response.json())

# Cell 3: Trigger jobs
import requests

# Trigger scan
response = requests.post(
    "http://localhost:8080/api/v1/jobs/scan",
    json={"stock_code": "600519", "params": {"days": 30}}
)
print(response.json())
```

## Monitoring

### Health Checks
```bash
# Local
curl http://localhost:8080/health

# Databricks (from notebook)
%sh curl http://localhost:8080/health
```

### Logs
```bash
# Docker
docker logs -f alpha-service

# Databricks
%sh tail -f /tmp/alpha.log
```

### Metrics
```bash
# Get system metrics
curl http://localhost:8080/api/v1/metrics
```

## Graceful Shutdown

The service handles SIGTERM and SIGINT signals gracefully:

```bash
# Send shutdown signal
kill -TERM <pid>

# Service will:
# 1. Stop accepting new requests
# 2. Complete in-flight requests
# 3. Stop background tasks
# 4. Clean up resources
# 5. Exit
```

Timeout: 30 seconds (configurable)

## Troubleshooting

### Service Won't Start
```bash
# Check logs
docker logs alpha-service

# Check configuration
export LOG_FORMAT=pretty
./bin/alpha-detector-service
```

### API Not Responding
```bash
# Check if port is bound
netstat -tuln | grep 8080

# Check service health
curl -v http://localhost:8080/health
```

### Background Tasks Not Running
```bash
# Verify run mode
echo $RUN_MODE

# Check scheduler logs
# Look for "Scheduled scan task triggered" messages
```

### High Memory Usage
```bash
# Check metrics
curl http://localhost:8080/api/v1/status

# Adjust worker concurrency
export WORKER_CONCURRENCY=1
```

## Migration from CLI

### Before (CLI Mode)
```bash
# Cron job
0 */1 * * * /path/to/alpha-detector scan-real
```

### After (Service Mode)
```bash
# Start service once
docker run -d \
  -e RUN_MODE=scheduler \
  -e SCAN_INTERVAL=1h \
  alpha-detector-service:latest
```

### Hybrid Approach
You can still use CLI commands while the service is running:

```bash
# Service runs in background
docker run -d --name alpha-service alpha-detector-service:latest

# Use CLI for one-off tasks
docker exec alpha-service /app/alpha-detector backtest --stock 600519
```

## Best Practices

1. **Use Environment Variables**: Configure via environment, not files
2. **Health Checks**: Always configure health check endpoints
3. **Graceful Shutdown**: Handle SIGTERM properly
4. **Resource Limits**: Set memory and CPU limits
5. **Logging**: Use structured JSON logging in production
6. **Monitoring**: Export metrics for Prometheus/Grafana
7. **Secrets**: Use secret management (Databricks Secrets, AWS Secrets Manager)
8. **Scaling**: Run multiple workers in worker mode for parallel processing

## Cost Optimization

### ARM64 Savings
- **x86 (m5.xlarge)**: $0.192/hour = $138.24/month
- **ARM64 (m6g.xlarge)**: $0.154/hour = $110.88/month
- **Savings**: 19.8% ($27.36/month per instance)

### Resource Optimization
- **API Mode**: 512MB RAM, 0.5 CPU
- **Scheduler Mode**: 256MB RAM, 0.25 CPU
- **Worker Mode**: 1GB RAM, 1 CPU
- **All Mode**: 1.5GB RAM, 1 CPU

## Support

For issues and questions:
- GitHub Issues: https://github.com/bingaigc/gp/issues
- Documentation: docs/README.md
- API Docs: http://localhost:8080/api/v1/docs (when enabled)
