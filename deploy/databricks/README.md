# Databricks Deployment Guide

## Overview

This directory contains templates and scripts for deploying the Alpha Detector Service on Databricks using custom Docker containers.

## Prerequisites

1. **Docker Registry**: Push your image to a registry accessible by Databricks
   - AWS ECR
   - Azure Container Registry
   - Docker Hub
   - Google Container Registry

2. **Databricks Secrets**: Create secrets for sensitive data
   ```bash
   databricks secrets create-scope alpha-detector
   databricks secrets put-secret --scope alpha-detector --key kimi-api-key
   databricks secrets put-secret --scope alpha-detector --key redis-url
   databricks secrets put-secret --scope alpha-detector --key postgres-url
   databricks secrets put-secret --scope alpha-detector --key kafka-brokers
   ```

3. **Container Registry Authentication** (if private registry):
   ```bash
   databricks secrets create-scope container-registry
   databricks secrets put-secret --scope container-registry --key username
   databricks secrets put-secret --scope container-registry --key password
   ```

## Building and Pushing Docker Image

### Option 1: Multi-arch build (recommended for ARM64 clusters)

```bash
# Build for both architectures
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f Dockerfile.service \
  -t your-registry/alpha-detector-service:latest \
  --push .
```

### Option 2: Single architecture

```bash
# Build ARM64 for cost savings (m6g instances)
docker buildx build \
  --platform linux/arm64 \
  -f Dockerfile.service \
  -t your-registry/alpha-detector-service:latest \
  --push .
```

### Push to Specific Registries

**AWS ECR:**
```bash
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin 123456789.dkr.ecr.us-east-1.amazonaws.com

docker tag alpha-detector-service:latest \
  123456789.dkr.ecr.us-east-1.amazonaws.com/alpha-detector-service:latest

docker push 123456789.dkr.ecr.us-east-1.amazonaws.com/alpha-detector-service:latest
```

**Azure ACR:**
```bash
az acr login --name myregistry

docker tag alpha-detector-service:latest \
  myregistry.azurecr.io/alpha-detector-service:latest

docker push myregistry.azurecr.io/alpha-detector-service:latest
```

## Deploying to Databricks

### Step 1: Upload keep_alive.py script

```bash
databricks fs cp keep_alive.py dbfs:/databricks/scripts/keep_alive.py
```

### Step 2: Update Job Template

Edit `databricks_job_template.json`:
- Replace `your-registry.azurecr.io/alpha-detector-service:latest` with your image URL
- Adjust environment variables as needed
- Update node_type_id for your preferred instance type

### Step 3: Create Databricks Job

```bash
databricks jobs create --json-file databricks_job_template.json
```

Or use the Databricks UI:
1. Go to Workflows → Jobs
2. Create Job
3. Select "JSON" and paste the template
4. Modify as needed
5. Save and Run

## Architecture on Databricks

```
┌─────────────────────────────────────┐
│   Databricks Job (Single Node)     │
│                                     │
│  ┌───────────────────────────────┐ │
│  │  Docker Container             │ │
│  │                               │ │
│  │  ┌─────────┐  ┌────────────┐ │ │
│  │  │   API   │  │ Scheduler  │ │ │
│  │  │  :8080  │  │  (Tasks)   │ │ │
│  │  └─────────┘  └────────────┘ │ │
│  │                               │ │
│  │  ┌────────────────────────┐  │ │
│  │  │  Alpha Detector Logic  │  │ │
│  │  └────────────────────────┘  │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │  keep_alive.py (Monitor)      │ │
│  └───────────────────────────────┘ │
└─────────────────────────────────────┘
```

## Environment Variables

Configure via `spark_env_vars` in the job template:

| Variable | Description | Default |
|----------|-------------|---------|
| `RUN_MODE` | Service mode (all\|api\|scheduler\|worker) | `all` |
| `API_PORT` | API server port | `8080` |
| `LOG_LEVEL` | Logging level (debug\|info\|warn\|error) | `info` |
| `LOG_FORMAT` | Log format (json\|pretty) | `json` |
| `SCAN_INTERVAL` | Scan interval (e.g., 1h, 30m) | `1h` |
| `PREDICT_INTERVAL` | Prediction interval | `6h` |
| `RUN_ON_STARTUP` | Run tasks on startup | `true` |
| `ALPHA_KIMI_API_KEY` | Kimi AI API key (from secrets) | - |
| `REDIS_URL` | Redis connection URL | - |
| `POSTGRES_URL` | PostgreSQL URL | - |
| `KAFKA_BROKERS` | Kafka broker list | - |

## Accessing the Service

### From Databricks Notebooks

```python
import requests

# Health check
response = requests.get("http://localhost:8080/health")
print(response.json())

# Trigger scan job
response = requests.post(
    "http://localhost:8080/api/v1/jobs/scan",
    json={"stock_code": "600519"}
)
print(response.json())

# Get system status
response = requests.get("http://localhost:8080/api/v1/status")
print(response.json())
```

### Available API Endpoints

- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /api/v1/status` - System status
- `GET /api/v1/version` - Version info
- `POST /api/v1/jobs/scan` - Trigger scan job
- `POST /api/v1/jobs/predict` - Trigger prediction
- `POST /api/v1/jobs/backtest` - Trigger backtest
- `GET /api/v1/signals` - Query signals
- `GET /api/v1/metrics` - Get metrics

## Monitoring

### Check Service Status

```python
# In a Databricks notebook
import requests
import json

def check_service():
    try:
        response = requests.get("http://localhost:8080/api/v1/status", timeout=5)
        return response.json()
    except Exception as e:
        return {"error": str(e)}

status = check_service()
print(json.dumps(status, indent=2))
```

### View Logs

Logs are available in:
1. Databricks Job UI → Run Details → Logs
2. Container stdout/stderr (JSON format)

## Cost Optimization

### Use ARM64 Instances (19.8% savings)

Update job template:
```json
{
  "node_type_id": "m6g.xlarge"  // ARM64
}
```

### Use Spot Instances

Already configured in template:
```json
{
  "aws_attributes": {
    "availability": "SPOT_WITH_FALLBACK"
  }
}
```

### Monthly Cost Comparison

| Instance | Type | Hourly | Monthly (24/7) | Savings |
|----------|------|--------|----------------|---------|
| m5.xlarge | x86_64 | $0.192 | $138.24 | - |
| m6g.xlarge | ARM64 | $0.154 | $110.88 | 19.8% |

## Troubleshooting

### Container fails to start

1. Check image URL is correct
2. Verify registry authentication
3. Ensure secrets are configured
4. Check Databricks has network access to registry

### Service not responding

1. Check health endpoint: `curl http://localhost:8080/health`
2. View container logs in Databricks UI
3. Verify RUN_MODE and environment variables
4. Check port 8080 is not blocked

### Job keeps restarting

1. Review `keep_alive.py` output
2. Check service health failures
3. Increase `max_failures` threshold if needed
4. Verify external dependencies (Redis, Postgres, Kafka)

## Advanced Configuration

### Running Multiple Modes

Deploy separate jobs for different modes:

**API-only job:**
```json
{
  "spark_env_vars": {
    "RUN_MODE": "api",
    ...
  }
}
```

**Scheduler-only job:**
```json
{
  "spark_env_vars": {
    "RUN_MODE": "scheduler",
    ...
  }
}
```

### Custom Intervals

Adjust scheduling:
```json
{
  "spark_env_vars": {
    "SCAN_INTERVAL": "30m",
    "PREDICT_INTERVAL": "3h",
    ...
  }
}
```

## Support

For issues or questions:
1. Check logs in Databricks UI
2. Review service documentation
3. Verify Docker image builds locally
4. Test API endpoints manually

## Files in This Directory

- `databricks_job_template.json` - Job configuration template
- `keep_alive.py` - Service monitoring script
- `README.md` - This deployment guide
