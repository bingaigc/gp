# Databricks ARM64 Deployment Guide

This guide provides step-by-step instructions for deploying the CAMP-I Alpha Signal Detection System on Databricks with ARM64 architecture using Docker containers.

## Important Notes

⚠️ **Databricks Limitations:**
- Databricks does NOT support `sudo`, `apt-get`, or local package installation
- Databricks does NOT support Docker daemon or `docker-compose`
- Init scripts cannot install system-level dependencies
- **The ONLY supported deployment method is using custom Docker images**

## Prerequisites

- Databricks workspace with cluster access
- Docker registry (Docker Hub, AWS ECR, Azure ACR, or GCR)
- Pre-built ARM64 Docker image of the application
- API keys (Kimi AI, if using real-time mode) stored in Databricks Secrets

## Table of Contents

1. [Build Docker Image](#build-docker-image)
2. [Push to Registry](#push-to-registry)
3. [Databricks Job Configuration](#databricks-job-configuration)
4. [Running the System](#running-the-system)
5. [Troubleshooting](#troubleshooting)
6. [Performance Optimization](#performance-optimization)

## Build Docker Image

### Step 1: Build ARM64 Image Locally or in CI/CD

On an ARM64 machine or using Docker Buildx:

```bash
# Clone repository
git clone https://github.com/bingaigc/gp.git
cd gp

# Build multi-arch image (supports both ARM64 and x86_64)
docker buildx build \
  --platform linux/arm64,linux/amd64 \
  -f Dockerfile.service \
  -t your-registry/alpha-detector-service:latest \
  --push .
```

Or build ARM64-only image:

```bash
# Build ARM64 image
docker build \
  --platform linux/arm64 \
  -f Dockerfile.service \
  -t your-registry/alpha-detector-service:arm64-latest \
  .
```

### Step 2: Verify Image

```bash
# Test locally
docker run --rm -e RUN_MODE=api -p 8080:8080 \
  your-registry/alpha-detector-service:latest

# Check health
curl http://localhost:8080/health
```

## Push to Registry

### Option 1: Docker Hub

```bash
# Login
docker login

# Push
docker push your-registry/alpha-detector-service:latest
```

### Option 2: AWS ECR (Recommended for Databricks on AWS)

```bash
# Authenticate
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin \
  123456789012.dkr.ecr.us-east-1.amazonaws.com

# Tag
docker tag your-registry/alpha-detector-service:latest \
  123456789012.dkr.ecr.us-east-1.amazonaws.com/alpha-detector-service:latest

# Push
docker push 123456789012.dkr.ecr.us-east-1.amazonaws.com/alpha-detector-service:latest
```

### Option 3: Azure ACR

```bash
# Login
az acr login --name yourregistry

# Push
docker push yourregistry.azurecr.io/alpha-detector-service:latest
```

### Method 1: Single-Container Job (Recommended)

**Use the long-running service mode** with the Docker image:

```json
{
    "name": "alpha-detector-service",
    "new_cluster": {
        "spark_version": "13.3.x-scala2.12",
        "node_type_id": "m6g.xlarge",
        "num_workers": 0,
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
        "LOG_FORMAT": "json",
        "SCAN_INTERVAL": "1h",
        "ALPHA_KIMI_API_KEY": "{{secrets/alpha-detector/kimi-api-key}}",
        "REDIS_URL": "{{secrets/alpha-detector/redis-url}}",
        "POSTGRES_URL": "{{secrets/alpha-detector/postgres-url}}",
        "KAFKA_BROKERS": "{{secrets/alpha-detector/kafka-brokers}}"
    },
    "schedule": {
        "quartz_cron_expression": "0 0 * * * ?",
        "timezone_id": "UTC",
        "pause_status": "UNPAUSED"
    }
}
```

### Method 2: Python API Call via Notebook

For interactive usage, call the service API from a Databricks notebook:

```python
# Databricks notebook
import requests
import json

# Service runs in the cluster via Docker image
# Call the API endpoint
API_URL = "http://localhost:8080/api/v1"

# Trigger scan job
response = requests.post(
    f"{API_URL}/jobs/scan",
    json={"stock_code": "600519", "params": {"days": 30}},
    headers={"Content-Type": "application/json"}
)

result = response.json()
print(json.dumps(result, indent=2))

# Check status
status = requests.get(f"{API_URL}/status").json()
print(f"Service status: {status['status']}")
```

## Setup Databricks Secrets

Before running, configure secrets for sensitive data:

```python
# Create secret scope (one-time setup)
from databricks.sdk import WorkspaceClient

w = WorkspaceClient()

# Create scope
w.secrets.create_scope(scope="alpha-detector")

# Add secrets via Databricks CLI:
# databricks secrets put --scope alpha-detector --key kimi-api-key
# databricks secrets put --scope alpha-detector --key redis-url
# databricks secrets put --scope alpha-detector --key postgres-url
```

Or via Databricks UI:
1. Go to **Settings** → **Secrets**
2. Create scope: `alpha-detector`
3. Add keys: `kimi-api-key`, `redis-url`, `postgres-url`, `kafka-brokers`

## Running the System

### Access Service Endpoints

Once the Docker container is running on Databricks:

```python
# Health check
%sh curl http://localhost:8080/health

# Get system status
%sh curl http://localhost:8080/api/v1/status

# Trigger jobs via API
%python
import requests

# Scan job
requests.post("http://localhost:8080/api/v1/jobs/scan", json={})

# Prediction job
requests.post("http://localhost:8080/api/v1/jobs/predict", json={
    "stock_code": "600519",
    "indicator": "MA5"
})

# Backtest job
requests.post("http://localhost:8080/api/v1/jobs/backtest", json={
    "stock_code": "600519",
    "days": 365
})
```

### Query Results via Delta Lake

Store and query signals in Delta Lake:

```python
from pyspark.sql.types import *

# Define schema
schema = StructType([
    StructField("stock_code", StringType(), True),
    StructField("stock_name", StringType(), True),
    StructField("signal_type", StringType(), True),
    StructField("total_score", DoubleType(), True),
    StructField("timestamp", TimestampType(), True),
    StructField("recommendation", StringType(), True)
])

# Read from service output (if writing to DBFS)
# Or call API to get signals
import requests
signals_response = requests.get("http://localhost:8080/api/v1/signals")
signals_data = signals_response.json()

# Create DataFrame
df = spark.createDataFrame(signals_data, schema)

# Write to Delta table
df.write.format("delta").mode("append").saveAsTable("alpha_signals")

# Query signals
spark.sql("""
    SELECT stock_code, stock_name, signal_type, total_score, timestamp
    FROM alpha_signals
    WHERE signal_type = 'STRONG_BUY'
    AND timestamp >= current_date() - INTERVAL 7 DAYS
    ORDER BY total_score DESC
""").show()
```

## Monitoring

### Health Checks

```python
import requests
import time

def check_service_health():
    try:
        response = requests.get("http://localhost:8080/health", timeout=5)
        if response.status_code == 200:
            health_data = response.json()
            print(f"✅ Service is healthy: {health_data}")
            return True
        else:
            print(f"❌ Service returned status {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ Service unreachable: {e}")
        return False

# Run health check
check_service_health()
```

### View Logs

```python
# In Databricks, view container logs
%sh
# Logs are available in Databricks cluster logs
# Or query via Spark if logs are written to DBFS
```

### Metrics Dashboard

```python
import requests
import pandas as pd

# Get metrics
metrics = requests.get("http://localhost:8080/api/v1/metrics").json()

# Convert to DataFrame
df_metrics = pd.DataFrame([metrics])
display(df_metrics)
```

### Common Issues

#### Issue 1: Docker Image Not Found

**Problem:** Job fails with "image not found" error.

**Solution:**
1. Verify image exists in registry:
   ```bash
   docker pull your-registry/alpha-detector-service:latest
   ```
2. Check registry credentials in Databricks Secrets
3. Ensure image tag matches job configuration

#### Issue 2: Service Won't Start

**Problem:** Container starts but service doesn't respond.

**Solution:**
1. Check environment variables are set correctly
2. Verify secrets are accessible
3. Review container logs in Databricks Job UI
4. Test image locally before deploying

#### Issue 3: API Key Issues

**Problem:** Service can't access Kimi AI API.

**Solution:**
1. Verify secret scope and key names:
   ```python
   dbutils.secrets.list(scope="alpha-detector")
   ```
2. Test secret access:
   ```python
   api_key = dbutils.secrets.get(scope="alpha-detector", key="kimi-api-key")
   print(f"Key starts with: {api_key[:10]}")
   ```
3. Update job configuration with correct secret paths

#### Issue 4: Network Connectivity

**Problem:** Service can't reach external APIs.

**Solution:**
1. Check Databricks network security groups
2. Verify egress rules allow HTTPS (port 443)
3. Test connectivity from notebook:
   ```python
   %sh curl -I https://api.moonshot.cn/v1
   ```

#### Issue 5: DBFS Permission Issues

**Problem:** Can't write to DBFS paths.

**Solution:**
1. Ensure paths use correct DBFS format: `/dbfs/path` not `dbfs:/path`
2. Create directories beforehand via notebook:
   ```python
   dbutils.fs.mkdirs("/alpha-detector/signals")
   dbutils.fs.mkdirs("/alpha-detector/data")
   ```

## Performance Optimization

### 1. Instance Type Selection

**ARM64 Instance Recommendations:**

| Workload | Instance Type | vCPU | Memory | Cost/Hour |
|----------|--------------|------|---------|-----------|
| Light | m6g.large | 2 | 8 GB | $0.077 |
| Medium | m6g.xlarge | 4 | 16 GB | $0.154 |
| Heavy | m6g.2xlarge | 8 | 32 GB | $0.308 |
| Intensive | m6g.4xlarge | 16 | 64 GB | $0.616 |

**Cost Savings:**
- m6g vs m5: **19.8% savings**
- c6g vs c5: **20% savings**  
- r6g vs r5: **20% savings**

### 2. Service Configuration

```json
{
    "spark_env_vars": {
        "RUN_MODE": "all",
        "API_PORT": "8080",
        "LOG_LEVEL": "info",
        "SCAN_INTERVAL": "1h",
        "WORKER_CONCURRENCY": "5",
        "GIN_MODE": "release"
    }
}
```

### 3. Auto-Scaling

Enable cluster auto-scaling for cost optimization:

```json
{
    "new_cluster": {
        "autoscale": {
            "min_workers": 1,
            "max_workers": 5
        }
    }
}
```

### 4. Monitoring Integration

```python
# Query service metrics
import requests
import time

def collect_metrics():
    metrics = requests.get("http://localhost:8080/api/v1/metrics").json()
    
    # Store in Delta for analysis
    df = spark.createDataFrame([{
        "timestamp": time.time(),
        **metrics
    }])
    
    df.write.format("delta").mode("append").saveAsTable("service_metrics")

# Schedule metric collection
while True:
    collect_metrics()
    time.sleep(60)  # Every minute
```

## Cost Analysis

### Monthly Cost Breakdown (24/7 Operation)

**Scenario 1: Single m6g.xlarge instance**
```
Instance: m6g.xlarge (ARM64)
Cost: $0.154/hour
Monthly: $0.154 × 730 hours = $112.42

vs x86 m5.xlarge: $0.192/hour = $140.16/month
Savings: $27.74/month (19.8%)
```

**Scenario 2: Auto-scaling (avg 2 workers)**
```
Driver: m6g.xlarge = $112.42/month
Workers: 2 × m6g.xlarge = $224.84/month
Total: $337.26/month

vs x86: $420.48/month
Savings: $83.22/month (19.8%)
```

**Additional Costs:**
- Docker Registry: ~$5/month (Docker Hub) or included (AWS ECR)
- Databricks Job scheduler: Included
- External services (Redis/PostgreSQL): Variable

**Total Monthly Cost: $342 - $450/month** (vs $430-$560 for x86)

## Best Practices

### 1. Use Pre-built Images

✅ Build once, deploy many times
✅ Faster job startup (<1 minute)
✅ Consistent environment

### 2. Leverage Databricks Secrets

✅ Never hardcode API keys
✅ Centralized secret management
✅ Audit trail for access

### 3. Enable Auto-scaling

✅ Cost optimization
✅ Handle variable workloads
✅ Automatic capacity management

### 4. Monitor Service Health

✅ Regular health checks
✅ Alert on failures
✅ Track metrics over time

### 5. Use Delta Lake for Results

✅ ACID transactions
✅ Time travel capabilities
✅ Efficient storage

## Summary

**Key Points:**
- ✅ Use custom Docker images (ONLY deployment method)
- ✅ NO sudo, apt-get, or local installs
- ✅ Configure via environment variables
- ✅ Leverage Databricks Secrets for credentials
- ✅ ARM64 provides 19.8% cost savings
- ✅ Service runs 24/7 with auto-restart
- ✅ API endpoints for job triggering
- ✅ Delta Lake integration for analytics

**Next Steps:**
1. Build and push Docker image to registry
2. Configure Databricks Secrets
3. Create Databricks Job with Docker image
4. Monitor via health endpoints
5. Query results from Delta Lake

For more details, see:
- `docs/SERVICE_DEPLOYMENT.md` - Service configuration
- `docs/QUICK_START_SERVICE.md` - Quick start guide
- `Dockerfile.service` - Docker image definition
```

### Query Signals

```python
# View recent signals
display(spark.sql("""
SELECT * 
FROM alpha_signals 
WHERE processed_date = current_date()
ORDER BY total_score DESC
LIMIT 10
"""))

# Analyze signal performance
display(spark.sql("""
SELECT 
    signal_type,
    COUNT(*) as count,
    AVG(total_score) as avg_score,
    MAX(total_score) as max_score
FROM alpha_signals
WHERE processed_date >= date_sub(current_date(), 30)
GROUP BY signal_type
ORDER BY avg_score DESC
"""))
```

## Monitoring and Alerting

### Prometheus Integration

```python
%sh
# Access Prometheus metrics
curl http://localhost:9090/metrics
```

### Create Dashboard

```python
# Create custom Databricks dashboard
from databricks.sdk import WorkspaceClient

w = WorkspaceClient()

dashboard = w.dashboards.create(
    name="Alpha Detector Performance",
    parent="folders/12345",
    tags=["alpha-detector", "monitoring"]
)
```

## Best Practices

1. **Use Databricks Secrets** for API keys
2. **Store data in DBFS** for persistence
3. **Use Delta tables** for signal history
4. **Schedule jobs** during market hours
5. **Monitor resource usage** to optimize costs
6. **Enable autoscaling** for variable workloads
7. **Use ARM64 instances** for cost efficiency
8. **Implement error handling** and retries
9. **Log to DBFS** for troubleshooting
10. **Test in dev workspace** before production

## Cost Optimization

### ARM64 vs x86 Cost Comparison

| Instance Type | vCPU | Memory | Hourly Cost | Monthly (24/7) |
|---------------|------|--------|-------------|----------------|
| m6g.xlarge (ARM64) | 4 | 16GB | $0.154 | ~$111 |
| m5.xlarge (x86) | 4 | 16GB | $0.192 | ~$139 |
| **Savings** | - | - | **19.8%** | **~$28/month** |

**Recommendation:** Use ARM64 instances (m6g family) for:
- 19.8% cost savings
- Similar performance
- Better energy efficiency

## Conclusion

This guide provides everything needed to deploy and run the CAMP-I Alpha Signal Detection System on Databricks with ARM64 architecture. For questions or issues, refer to:

- Main documentation: `README.md`
- API documentation: `docs/API.md`
- Troubleshooting: `docs/TROUBLESHOOTING.md`
- GitHub Issues: https://github.com/bingaigc/gp/issues

## Additional Resources

- [Databricks Runtime Release Notes](https://docs.databricks.com/release-notes/runtime/releases.html)
- [Databricks Jobs API](https://docs.databricks.com/dev-tools/api/latest/jobs.html)
- [Delta Lake Documentation](https://docs.delta.io/latest/index.html)
- [Go ARM64 Downloads](https://go.dev/dl/)
