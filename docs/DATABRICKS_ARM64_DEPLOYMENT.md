# Databricks ARM64 Deployment Guide

This guide provides step-by-step instructions for deploying the CAMP-I Alpha Signal Detection System on Databricks with ARM64 architecture.

## Prerequisites

- Databricks workspace with cluster access
- ARM64-compatible Databricks cluster (Runtime 13.0+ recommended)
- GitHub repository access
- API keys (Kimi AI, if using real-time mode)

## Table of Contents

1. [Cluster Configuration](#cluster-configuration)
2. [Installation Steps](#installation-steps)
3. [Configuration](#configuration)
4. [Running the System](#running-the-system)
5. [Troubleshooting](#troubleshooting)
6. [Performance Optimization](#performance-optimization)

## Cluster Configuration

### 1. Create ARM64 Cluster

Navigate to Databricks workspace and create a new cluster with the following specifications:

**Cluster Configuration:**
```
Databricks Runtime: 13.3 LTS or higher
Worker Type: ARM64 instance (e.g., m6g.xlarge, m6g.2xlarge)
Driver Type: ARM64 instance (e.g., m6g.xlarge)
Min Workers: 2
Max Workers: 8
Autoscaling: Enabled
```

**Advanced Options - Init Scripts:**
```bash
#!/bin/bash
# Install Go 1.21+ for ARM64
wget https://go.dev/dl/go1.21.5.linux-arm64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-arm64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# Install build essentials
sudo apt-get update
sudo apt-get install -y build-essential git make

# Verify installation
go version
```

### 2. Install Dependencies

Create a notebook and run:

```python
%sh
# Install system dependencies
sudo apt-get update && sudo apt-get install -y \
    build-essential \
    git \
    make \
    curl \
    wget

# Verify Go installation
go version

# Install Docker (optional, for container deployment)
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
```

## Installation Steps

### Step 1: Clone Repository

```python
%sh
cd /databricks/driver
git clone https://github.com/bingaigc/gp.git
cd gp
```

### Step 2: Build for ARM64

```python
%sh
cd /databricks/driver/gp

# Set ARM64 architecture
export GOARCH=arm64
export GOOS=linux

# Build the binary
make build

# Verify binary architecture
file bin/alpha-detector
# Expected output: bin/alpha-detector: ELF 64-bit LSB executable, ARM aarch64
```

### Step 3: Verify Installation

```python
%sh
cd /databricks/driver/gp
./bin/alpha-detector version
./bin/alpha-detector health
```

## Configuration

### 1. Create Configuration File

```python
%sh
cd /databricks/driver/gp

# Copy example configuration
cp configs/config.example.yaml configs/config.yaml

# Edit configuration (use sed or vi)
cat > configs/config.yaml << 'EOF'
data_source:
  primary: "eastmoney"
  timeout: 10s
  max_retry: 3

ai:
  provider: "kimi"
  model: "moonshot-v1-8k"
  base_url: "https://api.moonshot.cn/v1"
  timeout: 30s

risk:
  daily_max_loss: 50000.0
  max_signals_per_day: 20
  max_consecutive_losses: 3
  min_confidence: 0.6

pipeline:
  workers: 3
  batch_size: 20
  limit: 100

sinks:
  console:
    enabled: true
    pretty: true
  json_file:
    enabled: true
    output_dir: "/dbfs/alpha-detector/signals"
    pretty: true

filters:
  enable_basic: true
  enable_liquidity: true
  min_volume: 100000
  min_amount: 10000000.0

ml:
  historical:
    data_source: "eastmoney"
    default_days: 365
    storage_path: "/dbfs/alpha-detector/data/historical"
  
  prediction:
    default_model: "arima"
    confidence_threshold: 0.7
    cache_predictions: true

backtest:
  initial_capital: 1000000.0
  commission_rate: 0.0003
  slippage_rate: 0.001
  max_position_size: 0.15

monitoring:
  enabled: true
  prometheus_port: 9090
  metrics_interval: 60
EOF
```

### 2. Set Environment Variables

```python
# Set API keys in Databricks Secrets
# Go to Workspace > Settings > Secrets

# Create secret scope
dbutils.secrets.createScope("alpha-detector")

# Add secrets (via Databricks CLI or UI)
# dbutils.secrets.put(scope="alpha-detector", key="kimi-api-key", string_value="your-api-key")
```

```python
%sh
# Export environment variables
export ALPHA_KIMI_API_KEY=$(databricks secrets get --scope alpha-detector --key kimi-api-key)
export ALPHA_CONFIG_PATH=/databricks/driver/gp/configs/config.yaml
```

## Running the System

### Demo Mode (No API Key Required)

```python
%sh
cd /databricks/driver/gp
./bin/alpha-detector scan
```

### Real-Time Mode

```python
%sh
cd /databricks/driver/gp

# Set API key
export ALPHA_KIMI_API_KEY=$(databricks secrets get --scope alpha-detector --key kimi-api-key)

# Run real-time scan
./bin/alpha-detector scan-real
```

### ML Prediction Workflow

```python
%sh
cd /databricks/driver/gp

# Step 1: Collect historical data
./bin/alpha-detector collect --stock 600519 --days 365

# Step 2: Predict indicators
./bin/alpha-detector predict --stock 600519 --indicator MA5 --model arima

# Step 3: Run learning
./bin/alpha-detector learn --stock 600519 --indicator MA5

# Step 4: Analyze errors
./bin/alpha-detector analyze --stock 600519 --indicator MA5 --detailed
```

### Backtesting

```python
%sh
cd /databricks/driver/gp
./bin/alpha-detector backtest --stock 600519 --days 365 --capital 1000000
```

## Running as Scheduled Job

### Create Databricks Job

```python
from databricks.sdk import WorkspaceClient

w = WorkspaceClient()

job = w.jobs.create(
    name="Alpha-Detector-Daily-Scan",
    tasks=[
        {
            "task_key": "scan_market",
            "description": "Daily market scan",
            "notebook_task": {
                "notebook_path": "/Workspace/alpha-detector-notebook",
                "base_parameters": {
                    "command": "scan-real"
                }
            },
            "new_cluster": {
                "spark_version": "13.3.x-scala2.12",
                "node_type_id": "m6g.xlarge",
                "num_workers": 2
            },
            "timeout_seconds": 3600
        }
    ],
    schedule={
        "quartz_cron_expression": "0 0 9 * * ?",  # Daily at 9 AM
        "timezone_id": "Asia/Shanghai"
    }
)

print(f"Job created: {job.job_id}")
```

### Notebook Template

Create a notebook: `/Workspace/alpha-detector-notebook.py`

```python
# Databricks notebook source
# MAGIC %md
# MAGIC # Alpha Detector - Daily Scan

# COMMAND ----------
# Install and setup
%sh
cd /databricks/driver
if [ ! -d "gp" ]; then
  git clone https://github.com/bingaigc/gp.git
fi
cd gp
git pull
make build

# COMMAND ----------
# Configure environment
import os
dbutils.widgets.text("command", "scan-real", "Command to run")
command = dbutils.widgets.get("command")

# Set API key from secrets
os.environ['ALPHA_KIMI_API_KEY'] = dbutils.secrets.get(scope="alpha-detector", key="kimi-api-key")

# COMMAND ----------
# Run alpha detector
%sh
cd /databricks/driver/gp
export ALPHA_KIMI_API_KEY=$(databricks secrets get --scope alpha-detector --key kimi-api-key)
./bin/alpha-detector $command

# COMMAND ----------
# Save results to Delta table
from pyspark.sql.types import *
import json

# Read JSON signals
signal_files = dbutils.fs.ls("/dbfs/alpha-detector/signals")

signals = []
for file in signal_files:
    if file.name.endswith('.json'):
        with open(file.path.replace("dbfs:", "/dbfs"), 'r') as f:
            signals.append(json.load(f))

# Create DataFrame
schema = StructType([
    StructField("stock_code", StringType(), True),
    StructField("stock_name", StringType(), True),
    StructField("signal_type", StringType(), True),
    StructField("total_score", DoubleType(), True),
    StructField("timestamp", StringType(), True)
])

df = spark.createDataFrame(signals, schema)

# Write to Delta
df.write.format("delta").mode("append").save("/delta/alpha_signals")

print(f"Saved {len(signals)} signals to Delta table")
```

## Docker Deployment on Databricks

### Build ARM64 Docker Image

```python
%sh
cd /databricks/driver/gp

# Create Dockerfile for ARM64
cat > Dockerfile.arm64 << 'EOF'
FROM arm64v8/golang:1.21-alpine AS builder

WORKDIR /app
COPY . .

# Build for ARM64
RUN GOARCH=arm64 GOOS=linux go build -o bin/alpha-detector ./cmd/alpha

FROM arm64v8/alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/bin/alpha-detector .
COPY --from=builder /app/configs ./configs

EXPOSE 8080 9090

ENTRYPOINT ["./alpha-detector"]
CMD ["serve"]
EOF

# Build image
docker build -f Dockerfile.arm64 -t alpha-detector:arm64 .
```

### Run with Docker Compose

```python
%sh
cd /databricks/driver/gp

# Create docker-compose override for ARM64
cat > docker-compose.arm64.yml << 'EOF'
version: '3.8'

services:
  alpha-detector:
    image: alpha-detector:arm64
    platform: linux/arm64
    build:
      context: .
      dockerfile: Dockerfile.arm64
    environment:
      - ALPHA_KIMI_API_KEY=${ALPHA_KIMI_API_KEY}
    ports:
      - "8080:8080"
      - "9090:9090"
    volumes:
      - /dbfs/alpha-detector/data:/app/data
      - /dbfs/alpha-detector/signals:/app/output/signals
    networks:
      - alpha-network

networks:
  alpha-network:
    driver: bridge
EOF

# Start services
docker-compose -f docker-compose.yml -f docker-compose.arm64.yml up -d
```

## Troubleshooting

### Issue 1: Go Not Found

```python
%sh
# Reinstall Go for ARM64
wget https://go.dev/dl/go1.21.5.linux-arm64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-arm64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### Issue 2: Permission Denied on DBFS

```python
%sh
# Create directories with proper permissions
sudo mkdir -p /dbfs/alpha-detector/signals
sudo mkdir -p /dbfs/alpha-detector/data
sudo chmod -R 777 /dbfs/alpha-detector
```

### Issue 3: Binary Architecture Mismatch

```python
%sh
# Verify and rebuild for correct architecture
cd /databricks/driver/gp
file bin/alpha-detector

# If wrong architecture, rebuild
export GOARCH=arm64
export GOOS=linux
go clean
make build
```

### Issue 4: Network Connectivity

```python
%sh
# Test connectivity to APIs
curl -I https://api.moonshot.cn/v1
curl -I https://push2.eastmoney.com

# Check DNS
nslookup api.moonshot.cn
```

## Performance Optimization

### 1. Memory Configuration

```python
# Increase driver memory
spark.conf.set("spark.driver.memory", "8g")
spark.conf.set("spark.executor.memory", "8g")
```

### 2. Parallel Processing

```python
%sh
cd /databricks/driver/gp

# Use more workers for parallel processing
# Edit config.yaml
sed -i 's/workers: 3/workers: 8/' configs/config.yaml
```

### 3. Caching

```python
%sh
# Enable Redis for caching (optional)
docker run -d --name redis-arm64 --platform linux/arm64 \
  -p 6379:6379 arm64v8/redis:alpine
```

### 4. Resource Monitoring

```python
%sh
# Monitor resource usage
top -n 1
free -h
df -h /dbfs

# Monitor alpha-detector process
ps aux | grep alpha-detector
```

## Integration with Delta Lake

### Save Signals to Delta Table

```python
from pyspark.sql.functions import *
from delta.tables import *

# Create Delta table for signals
spark.sql("""
CREATE TABLE IF NOT EXISTS alpha_signals (
    stock_code STRING,
    stock_name STRING,
    signal_type STRING,
    total_score DOUBLE,
    capital_score DOUBLE,
    technical_score DOUBLE,
    valuation_score DOUBLE,
    sector_score DOUBLE,
    institutional_score DOUBLE,
    risk_level STRING,
    timestamp TIMESTAMP,
    processed_date DATE
)
USING DELTA
PARTITIONED BY (processed_date)
LOCATION '/delta/alpha_signals'
""")

# Load and append new signals
import json
from datetime import datetime

signal_files = dbutils.fs.ls("/dbfs/alpha-detector/signals")

for file in signal_files:
    if file.name.endswith('.json'):
        with open(file.path.replace("dbfs:", "/dbfs"), 'r') as f:
            data = json.load(f)
            data['processed_date'] = datetime.now().date()
            
            df = spark.createDataFrame([data])
            df.write.format("delta").mode("append").saveAsTable("alpha_signals")
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
