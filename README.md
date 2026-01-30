# Kibana Prometheus Exporter

A lightweight, secure Prometheus exporter for Kibana metrics, written in Go.

## Features

- Scrapes Kibana's `/api/status` endpoint for metrics
- Minimal dependencies, no known vulnerabilities
- Runs as a non-root user in a scratch container
- Supports Kibana 7.x and 8.x
- TLS and basic authentication support
- Kubernetes-ready with health/readiness probes
- Compatible with Prometheus ServiceMonitor

## Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `kibana_up` | Gauge | Was the last scrape of Kibana successful (1/0) |
| `kibana_status_overall` | Gauge | Overall status (1=green, 0.5=yellow, 0=red) |
| `kibana_status_core` | Gauge | Core service status by name |
| `kibana_status_elasticsearch` | Gauge | Elasticsearch connection status |
| `kibana_heap_total_bytes` | Gauge | Total heap size |
| `kibana_heap_used_bytes` | Gauge | Used heap size |
| `kibana_memory_resident_set_bytes` | Gauge | Resident set size |
| `kibana_event_loop_delay_seconds` | Gauge | Event loop delay |
| `kibana_requests_total` | Counter | Total requests by status |
| `kibana_response_time_seconds` | Gauge | Response time (avg/max) |
| `kibana_concurrent_connections_total` | Gauge | Concurrent connections |
| `kibana_process_uptime_seconds` | Gauge | Process uptime |
| `kibana_os_cpu_percent` | Gauge | OS CPU usage |
| `kibana_os_load_average_*` | Gauge | Load averages (1m/5m/15m) |
| `kibana_os_memory_*_bytes` | Gauge | OS memory (total/free/used) |
| `kibana_scrape_duration_seconds` | Gauge | Scrape duration |

## Quick Start

### Binary

```bash
# Download and run
./kibana-exporter --kibana-url=http://localhost:5601

# With authentication
./kibana-exporter \
  --kibana-url=https://kibana.example.com \
  --kibana-username=elastic \
  --kibana-password=changeme \
  --insecure-skip-verify
```

### Docker

```bash
docker run -d \
  -p 9684:9684 \
  -e KIBANA_URL=http://kibana:5601 \
  rahulnutakki/kibana-prometheus-exporter:latest
```

### Kubernetes

```bash
kubectl apply -f deploy/kubernetes/deployment.yaml
kubectl apply -f deploy/kubernetes/servicemonitor.yaml
```

## Configuration

### Command Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--listen-address` | `:9684` | Address to listen on |
| `--metrics-path` | `/metrics` | Path for metrics endpoint |
| `--kibana-url` | `http://localhost:5601` | Kibana URL |
| `--kibana-username` | (empty) | Basic auth username |
| `--kibana-password` | (empty) | Basic auth password |
| `--timeout` | `10s` | Request timeout |
| `--insecure-skip-verify` | `false` | Skip TLS verification |
| `--log-level` | `info` | Log level (debug/info/warn/error) |
| `--log-format` | `text` | Log format (text/json) |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `KIBANA_URL` | Overrides `--kibana-url` |
| `KIBANA_USERNAME` | Overrides `--kibana-username` |
| `KIBANA_PASSWORD` | Overrides `--kibana-password` |

## Endpoints

| Endpoint | Description |
|----------|-------------|
| `/` | Landing page with links |
| `/metrics` | Prometheus metrics |
| `/health` | Liveness probe (always returns 200) |
| `/ready` | Readiness probe (checks Kibana connectivity) |

## Security

- Runs as non-root user (65534:65534)
- Uses scratch base image (no shell, minimal attack surface)
- Read-only root filesystem
- All capabilities dropped
- No known CVEs in dependencies

### Vulnerability Scanning

```bash
# Scan Go dependencies
govulncheck ./...

# Scan container image
trivy image rahulnutakki/kibana-prometheus-exporter:latest
```

## Building

### Prerequisites

- Go 1.22+
- Docker (for container builds)

### Build Commands

```bash
# Build binary
make build

# Run locally
make run

# Build Docker image
make docker

# Run tests
make test

# Check for vulnerabilities
make scan
```

## Prometheus Configuration

### Static Config

```yaml
scrape_configs:
  - job_name: 'kibana'
    static_configs:
      - targets: ['kibana-exporter:9684']
```

### ServiceMonitor (Prometheus Operator)

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: kibana-exporter
  labels:
    release: prometheus
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: kibana-prometheus-exporter
  endpoints:
    - port: metrics
      interval: 30s
```

## Grafana Dashboard

A sample Grafana dashboard is available in `deploy/grafana/dashboard.json`.

Import it via Grafana UI or use a ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kibana-dashboard
  labels:
    grafana_dashboard: "1"
data:
  kibana.json: |
    # Dashboard JSON here
```

## Troubleshooting

### Exporter can't connect to Kibana

1. Check Kibana URL is correct and accessible
2. Verify authentication credentials if required
3. Check `/ready` endpoint for specific errors:
   ```bash
   curl http://exporter:9684/ready
   ```

### No metrics returned

1. Check exporter logs for errors
2. Verify Kibana's `/api/status` endpoint is accessible:
   ```bash
   curl -u user:pass http://kibana:5601/api/status
   ```

### Missing OS metrics

Some Kibana deployments (especially containerized) may not expose all OS metrics. This is expected behavior.

## License

MIT License - see [LICENSE](LICENSE)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting: `make test lint`
5. Submit a pull request
