# Kubernetes Deployment Guide

## Prerequisites

- Kubernetes cluster (v1.20+)
- kubectl configured
- Helm 3.x (optional, for easier deployment)
- Docker registry access

## Quick Start

### 1. Create Namespace
```bash
kubectl create namespace trading
```

### 2. Create Secrets
```bash
# Create API key secret
kubectl create secret generic alpha-secrets \
  --from-literal=kimi-api-key='your-actual-api-key' \
  -n trading
```

### 3. Deploy Application
```bash
# Apply RBAC
kubectl apply -f deploy/kubernetes/rbac.yaml

# Deploy application
kubectl apply -f deploy/kubernetes/deployment.yaml
```

### 4. Verify Deployment
```bash
# Check pods
kubectl get pods -n trading

# Check services
kubectl get svc -n trading

# View logs
kubectl logs -f deployment/alpha-detector -n trading
```

## Scaling

### Manual Scaling
```bash
kubectl scale deployment alpha-detector --replicas=5 -n trading
```

### Auto-Scaling (HPA)
The HorizontalPodAutoscaler is automatically configured:
- Min replicas: 2
- Max replicas: 10
- CPU target: 70%
- Memory target: 80%
- Custom metric: signals_generated_total

## Monitoring

### Prometheus Integration
Service monitor is configured for automatic scraping:
```bash
kubectl get servicemonitor -n trading
```

### View Metrics
```bash
# Port-forward to access metrics
kubectl port-forward svc/alpha-detector-service 9090:9090 -n trading

# Access metrics at http://localhost:9090/metrics
```

## Configuration Updates

### Update ConfigMap
```bash
# Edit configmap
kubectl edit configmap alpha-config -n trading

# Restart pods to pick up changes
kubectl rollout restart deployment/alpha-detector -n trading
```

### Update Secrets
```bash
kubectl delete secret alpha-secrets -n trading
kubectl create secret generic alpha-secrets \
  --from-literal=kimi-api-key='new-api-key' \
  -n trading
kubectl rollout restart deployment/alpha-detector -n trading
```

## High Availability

### Multi-Zone Deployment
```yaml
# Add to deployment spec
spec:
  replicas: 3
  template:
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - alpha-detector
              topologyKey: topology.kubernetes.io/zone
```

### Resource Requests/Limits
Current configuration:
- CPU request: 500m, limit: 2000m
- Memory request: 512Mi, limit: 2Gi

Adjust based on workload:
```bash
kubectl set resources deployment alpha-detector \
  --requests=cpu=1000m,memory=1Gi \
  --limits=cpu=4000m,memory=4Gi \
  -n trading
```

## Persistent Storage

### PVC Configuration
10Gi storage for historical data and models:
```bash
kubectl get pvc -n trading
```

### Backup
```bash
# Create snapshot (if supported)
kubectl create snapshot alpha-data-snapshot \
  --pvc alpha-data-pvc \
  -n trading
```

## Troubleshooting

### Pod Not Starting
```bash
# Describe pod
kubectl describe pod <pod-name> -n trading

# Check events
kubectl get events -n trading --sort-by='.lastTimestamp'
```

### Health Check Failures
```bash
# Test liveness probe
kubectl exec -it <pod-name> -n trading -- curl http://localhost:8080/health

# Test readiness probe
kubectl exec -it <pod-name> -n trading -- curl http://localhost:8080/ready
```

### Resource Issues
```bash
# Check resource usage
kubectl top pods -n trading
kubectl top nodes
```

## Integration with Service Mesh (Optional)

### Istio Integration
```yaml
# Add to deployment annotations
spec:
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "true"
```

### Traffic Management
```yaml
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: alpha-detector-vs
spec:
  hosts:
  - alpha-detector-service
  http:
  - route:
    - destination:
        host: alpha-detector-service
      weight: 90
    - destination:
        host: alpha-detector-service-canary
      weight: 10
```

## Production Checklist

- [ ] API keys securely stored in secrets
- [ ] Resource limits configured
- [ ] HPA enabled and tested
- [ ] Persistent storage configured
- [ ] Monitoring and alerting setup
- [ ] Backup strategy in place
- [ ] Multi-zone deployment (HA)
- [ ] Network policies configured
- [ ] Pod security policies applied
- [ ] Log aggregation configured
- [ ] Disaster recovery tested
