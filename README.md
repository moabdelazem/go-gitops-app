# Go GitOps App

A Go application designed to demonstrate **Kubernetes Horizontal Pod Autoscaler (HPA)** behavior and GitOps practices. This project helps me understand how Kubernetes automatically scales workloads based on CPU utilization.

# Purpose

This project was created to:
- Understand Kubernetes HPA scaling behavior in practice
- Demonstrate GitOps workflows with Kustomize
- Provide a controllable stress endpoint to trigger auto-scaling
- Observe how pods scale up under load and scale down when idle

## Features

- **REST API** with health checks and Prometheus metrics
- **Stress Endpoint** - Generates CPU load to trigger HPA
- **Kustomize** - Environment-specific deployments (dev/prod)
- **HPA Configuration** - Scales from 1 to 10 pods at 50% CPU
- **k6 Load Testing** - Scripts to trigger and observe scaling

## Quick Start

### Prerequisites

- Go 
- Docker
- Kubernetes cluster (minikube, kind, or cloud)
- kubectl
- k6 (for load testing)

### Run Locally

```bash
# Clone the repository
git clone https://github.com/moabdelazem/go-gitops-app.git
cd go-gitops-app

# Run the application
make run

# Or with Docker
make docker-build
make docker-run

docker compose up
```

### Deploy to Kubernetes

```bash
# Deploy to development environment
kubectl apply -k k8s/overlays/dev

# Deploy to production environment
kubectl apply -k k8s/overlays/production

# Check deployment status
kubectl get pods -n go-gitops-dev
kubectl get hpa -n go-gitops-dev
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Welcome message with version |
| `/health` | GET | Health check for K8s probes |
| `/stress` | GET | CPU stress test endpoint |
| `/metrics` | GET | Prometheus metrics |

### Stress Endpoint

The `/stress` endpoint generates CPU load to trigger HPA scaling.

```bash
# Default: 2 seconds, all CPU cores
curl http://localhost:8080/stress

# Custom duration and workers
curl "http://localhost:8080/stress?duration=10s&workers=4"
```

**Parameters:**
- `duration` - Stress duration (1s-30s, default: 2s)
- `workers` - Number of CPU workers (default: number of cores)

## Project Structure

```
go-gitops-app/
├── cmd/                      # Application entrypoint
├── internal/
│   ├── handlers/             # HTTP handlers
│   └── middleware/           # Logging, recovery middleware
├── pkg/
│   ├── logger/               # Structured logging
│   ├── metrics/              # Prometheus metrics
│   └── response/             # JSON response helpers
├── k8s/
│   ├── base/                 # Base Kubernetes manifests
│   │   ├── deployment.yml
│   │   ├── service.yml
│   │   ├── configmap.yml
│   │   ├── hpa.yml
│   │   └── kustomization.yaml
│   └── overlays/             # Environment-specific configs
│       ├── dev/
│       └── production/
├── tests/
│   └── load/                 # k6 load testing scripts
├── Dockerfile
├── Makefile
└── docker-compose.yml
```

## Understanding HPA Scaling

### How It Works

1. **HPA monitors CPU usage** of pods via metrics-server
2. **When CPU exceeds 50%**, HPA creates additional replicas
3. **Load is distributed** across all pods
4. **When load decreases**, HPA scales down to minimum replicas

### HPA Configuration

```yaml
# k8s/base/hpa.yml
spec:
  minReplicas: 1
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 50
```

### Trigger Scaling with Load Test

```bash
# Port forward the service
kubectl port-forward svc/go-gitops-app 8080:80 -n go-gitops-dev

# Run k6 load test (in another terminal)
k6 run tests/load/stress-test.js

# Watch HPA scaling (in another terminal)
kubectl get hpa go-gitops-app -n go-gitops-dev -w

# Watch pods scaling
kubectl get pods -n go-gitops-dev -w
```

### Expected Behavior

| Stage | Duration | Virtual Users | Expected Pods |
|-------|----------|---------------|---------------|
| Ramp up | 30s | 0 → 10 | 1 → 2 |
| Sustained | 2m | 10 → 20 | 2 → 5 |
| Peak | 1m | 20 → 30 | 5 → 10 |
| Ramp down | 30s | 30 → 0 | 10 → 1 (gradual) |

> **Note:** Scale-down is slower than scale-up (by design) to prevent thrashing.

## Kustomize Overlays

This project uses Kustomize for environment-specific deployments.

### Development

```bash
kubectl apply -k k8s/overlays/dev
```
- Namespace: `go-gitops-dev`
- Log Level: `debug`

### Production

```bash
kubectl apply -k k8s/overlays/production
```
- Namespace: `go-gitops-prod`
- Log Level: `info`

### Preview Manifests

```bash
# See what will be applied without applying
kubectl kustomize k8s/overlays/dev
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
