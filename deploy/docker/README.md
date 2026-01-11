# Docker Compose Deployment Guide

## Quick Start

### Prerequisites
- Docker Engine 20.10+
- Docker Compose 2.0+  
- 8GB+ RAM, 20GB+ Disk

### One-Command Startup
```bash
docker-compose up -d && docker-compose logs -f
```

### Access Services
- **API**: http://localhost:8080
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9091
- **Jaeger**: http://localhost:16686

See full documentation for details.
