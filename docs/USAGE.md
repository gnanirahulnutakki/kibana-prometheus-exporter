# Kibana Prometheus Exporter - Usage Guide

This guide provides detailed instructions for deploying and using the Kibana Prometheus Exporter.

## Table of Contents

1. [Overview](#overview)
2. [Installation Methods](#installation-methods)
3. [Kubernetes Deployment](#kubernetes-deployment)
4. [Helm Chart Integration](#helm-chart-integration)
5. [Sidecar Deployment](#sidecar-deployment)
6. [Prometheus Configuration](#prometheus-configuration)
7. [Grafana Dashboard](#grafana-dashboard)
8. [Alerting Rules](#alerting-rules)

## Overview

The Kibana Prometheus Exporter collects metrics from Kibana's `/api/status` endpoint and exposes them in Prometheus format. It's designed for:

- Monitoring Kibana health and performance
- Tracking heap memory and event loop metrics
- Observing request patterns and response times
- Alerting on Kibana availability issues

## Installation Methods

### Method 1: Standalone Deployment

Deploy as a separate pod that scrapes Kibana:

```bash
kubectl apply -f deploy/kubernetes/deployment.yaml
kubectl apply -f deploy/kubernetes/servicemonitor.yaml
```

Edit the deployment to point to your Kibana instance:

```yaml
env:
  - name: KIBANA_URL
    value: "http://kibana:5601"
```

### Method 2: Sidecar Container

Add as a sidecar to your Kibana pod for network-level access:

```yaml
spec:
  containers:
    - name: kibana
      image: docker.elastic.co/kibana/kibana:8.11.0
      ports:
        - containerPort: 5601

    - name: exporter
      image: rahulnutakki/kibana-prometheus-exporter:latest
      args:
        - --kibana-url=http://localhost:5601
        - --log-level=info
      ports:
        - name: metrics
          containerPort: 9684
      resources:
        requests:
          cpu: 10m
          memory: 32Mi
        limits:
          cpu: 100m
          memory: 64Mi
```

### Method 3: Docker Compose

```yaml
version: '3.8'
services:
  kibana:
    image: docker.elastic.co/kibana/kibana:8.11.0
    ports:
      - "5601:5601"
    environment:
      ELASTICSEARCH_HOSTS: '["http://elasticsearch:9200"]'

  kibana-exporter:
    image: rahulnutakki/kibana-prometheus-exporter:latest
    ports:
      - "9684:9684"
    environment:
      KIBANA_URL: http://kibana:5601
    depends_on:
      - kibana
```

## Kubernetes Deployment

### Basic Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kibana-exporter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kibana-exporter
  template:
    metadata:
      labels:
        app: kibana-exporter
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9684"
    spec:
      containers:
        - name: exporter
          image: rahulnutakki/kibana-prometheus-exporter:latest
          env:
            - name: KIBANA_URL
              value: "http://kibana:5601"
          ports:
            - containerPort: 9684
          livenessProbe:
            httpGet:
              path: /health
              port: 9684
          readinessProbe:
            httpGet:
              path: /ready
              port: 9684
```

### With Authentication

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kibana-exporter-auth
type: Opaque
stringData:
  username: elastic
  password: changeme
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kibana-exporter
spec:
  template:
    spec:
      containers:
        - name: exporter
          image: rahulnutakki/kibana-prometheus-exporter:latest
          env:
            - name: KIBANA_URL
              value: "https://kibana:5601"
            - name: KIBANA_USERNAME
              valueFrom:
                secretKeyRef:
                  name: kibana-exporter-auth
                  key: username
            - name: KIBANA_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: kibana-exporter-auth
                  key: password
          args:
            - --insecure-skip-verify
```

## Helm Chart Integration

### common-services Integration

Add the exporter as a sidecar in your Kibana values:

```yaml
kibana:
  enabled: true
  extraContainers:
    - name: metrics-exporter
      image: rahulnutakki/kibana-prometheus-exporter:latest
      args:
        - --kibana-url=http://localhost:5601
        - --log-level=info
      ports:
        - name: metrics
          containerPort: 9684
          protocol: TCP
      resources:
        requests:
          cpu: 10m
          memory: 32Mi
        limits:
          cpu: 100m
          memory: 64Mi
      livenessProbe:
        httpGet:
          path: /health
          port: metrics
        initialDelaySeconds: 10
      readinessProbe:
        httpGet:
          path: /ready
          port: metrics
        initialDelaySeconds: 10
```

### ServiceMonitor for Sidecar

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: kibana-metrics
  labels:
    release: prometheus
spec:
  selector:
    matchLabels:
      app: kibana
  endpoints:
    - port: metrics
      interval: 30s
      path: /metrics
```

## Prometheus Configuration

### Static Configuration

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'kibana'
    static_configs:
      - targets: ['kibana-exporter:9684']
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
        replacement: 'kibana'
```

### Kubernetes SD

```yaml
scrape_configs:
  - job_name: 'kibana'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: kibana-exporter
      - source_labels: [__meta_kubernetes_namespace]
        target_label: namespace
```

## Grafana Dashboard

### Dashboard JSON

Create a ConfigMap for the Grafana sidecar:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kibana-dashboard
  labels:
    grafana_dashboard: "1"
data:
  kibana-dashboard.json: |
    {
      "annotations": {"list": []},
      "editable": true,
      "fiscalYearStartMonth": 0,
      "graphTooltip": 0,
      "id": null,
      "links": [],
      "panels": [
        {
          "datasource": {"type": "prometheus", "uid": "${datasource}"},
          "fieldConfig": {
            "defaults": {"mappings": [], "thresholds": {"mode": "absolute", "steps": [{"color": "red", "value": null}, {"color": "green", "value": 1}]}},
            "overrides": []
          },
          "gridPos": {"h": 4, "w": 4, "x": 0, "y": 0},
          "id": 1,
          "options": {"colorMode": "value", "graphMode": "none", "justifyMode": "auto", "orientation": "auto", "reduceOptions": {"calcs": ["lastNotNull"], "fields": "", "values": false}, "textMode": "auto"},
          "pluginVersion": "10.0.0",
          "targets": [{"expr": "kibana_up{namespace=~\"$namespace\"}", "refId": "A"}],
          "title": "Kibana Up",
          "type": "stat"
        },
        {
          "datasource": {"type": "prometheus", "uid": "${datasource}"},
          "fieldConfig": {
            "defaults": {"color": {"mode": "palette-classic"}, "custom": {"axisBorderShow": false, "axisCenteredZero": false, "axisColorMode": "text", "axisLabel": "", "axisPlacement": "auto", "barAlignment": 0, "drawStyle": "line", "fillOpacity": 10, "gradientMode": "none", "hideFrom": {"legend": false, "tooltip": false, "viz": false}, "insertNulls": false, "lineInterpolation": "linear", "lineWidth": 1, "pointSize": 5, "scaleDistribution": {"type": "linear"}, "showPoints": "never", "spanNulls": false, "stacking": {"group": "A", "mode": "none"}, "thresholdsStyle": {"mode": "off"}}, "mappings": [], "thresholds": {"mode": "absolute", "steps": [{"color": "green", "value": null}]}, "unit": "bytes"},
            "overrides": []
          },
          "gridPos": {"h": 8, "w": 12, "x": 0, "y": 4},
          "id": 2,
          "options": {"legend": {"calcs": [], "displayMode": "list", "placement": "bottom", "showLegend": true}, "tooltip": {"mode": "multi", "sort": "none"}},
          "targets": [
            {"expr": "kibana_heap_total_bytes{namespace=~\"$namespace\"}", "legendFormat": "Total", "refId": "A"},
            {"expr": "kibana_heap_used_bytes{namespace=~\"$namespace\"}", "legendFormat": "Used", "refId": "B"}
          ],
          "title": "Heap Memory",
          "type": "timeseries"
        },
        {
          "datasource": {"type": "prometheus", "uid": "${datasource}"},
          "fieldConfig": {
            "defaults": {"color": {"mode": "palette-classic"}, "custom": {"axisBorderShow": false, "axisCenteredZero": false, "axisColorMode": "text", "axisLabel": "", "axisPlacement": "auto", "barAlignment": 0, "drawStyle": "line", "fillOpacity": 10, "gradientMode": "none", "hideFrom": {"legend": false, "tooltip": false, "viz": false}, "insertNulls": false, "lineInterpolation": "linear", "lineWidth": 1, "pointSize": 5, "scaleDistribution": {"type": "linear"}, "showPoints": "never", "spanNulls": false, "stacking": {"group": "A", "mode": "none"}, "thresholdsStyle": {"mode": "off"}}, "mappings": [], "thresholds": {"mode": "absolute", "steps": [{"color": "green", "value": null}]}, "unit": "s"},
            "overrides": []
          },
          "gridPos": {"h": 8, "w": 12, "x": 12, "y": 4},
          "id": 3,
          "options": {"legend": {"calcs": [], "displayMode": "list", "placement": "bottom", "showLegend": true}, "tooltip": {"mode": "multi", "sort": "none"}},
          "targets": [
            {"expr": "kibana_response_time_seconds{namespace=~\"$namespace\", quantile=\"avg\"}", "legendFormat": "Avg", "refId": "A"},
            {"expr": "kibana_response_time_seconds{namespace=~\"$namespace\", quantile=\"max\"}", "legendFormat": "Max", "refId": "B"}
          ],
          "title": "Response Time",
          "type": "timeseries"
        }
      ],
      "refresh": "30s",
      "schemaVersion": 38,
      "tags": ["kibana", "elastic"],
      "templating": {
        "list": [
          {"current": {"selected": false, "text": "Prometheus", "value": "prometheus"}, "hide": 0, "includeAll": false, "multi": false, "name": "datasource", "options": [], "query": "prometheus", "refresh": 1, "regex": "", "skipUrlSync": false, "type": "datasource"},
          {"allValue": ".*", "current": {"selected": true, "text": "All", "value": "$__all"}, "datasource": {"type": "prometheus", "uid": "${datasource}"}, "definition": "label_values(kibana_up, namespace)", "hide": 0, "includeAll": true, "multi": false, "name": "namespace", "options": [], "query": {"query": "label_values(kibana_up, namespace)", "refId": "StandardVariableQuery"}, "refresh": 2, "regex": "", "skipUrlSync": false, "sort": 1, "type": "query"}
        ]
      },
      "time": {"from": "now-1h", "to": "now"},
      "timepicker": {},
      "timezone": "",
      "title": "Kibana Metrics",
      "uid": "kibana-metrics",
      "version": 1,
      "weekStart": ""
    }
```

## Alerting Rules

### PrometheusRule

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: kibana-alerts
  labels:
    release: prometheus
spec:
  groups:
    - name: kibana
      rules:
        - alert: KibanaDown
          expr: kibana_up == 0
          for: 5m
          labels:
            severity: critical
          annotations:
            summary: "Kibana is down"
            description: "Kibana has been unreachable for more than 5 minutes."

        - alert: KibanaStatusDegraded
          expr: kibana_status_overall < 1
          for: 10m
          labels:
            severity: warning
          annotations:
            summary: "Kibana status is degraded"
            description: "Kibana overall status is not green."

        - alert: KibanaHighHeapUsage
          expr: kibana_heap_used_bytes / kibana_heap_size_limit_bytes > 0.9
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "Kibana heap usage is high"
            description: "Kibana heap usage is above 90%."

        - alert: KibanaHighResponseTime
          expr: kibana_response_time_seconds{quantile="avg"} > 2
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "Kibana response time is high"
            description: "Kibana average response time is above 2 seconds."
```

## Troubleshooting

### Common Issues

**1. Exporter reports kibana_up=0**

Check connectivity:
```bash
kubectl exec -it kibana-exporter-xxx -- wget -qO- http://kibana:5601/api/status
```

Check exporter logs:
```bash
kubectl logs kibana-exporter-xxx
```

**2. Missing metrics**

Not all Kibana deployments expose all metrics. Container deployments may not have OS-level metrics.

**3. Authentication errors**

Ensure credentials are correctly set:
```bash
kubectl get secret kibana-exporter-auth -o jsonpath='{.data.password}' | base64 -d
```

**4. TLS/SSL issues**

Use `--insecure-skip-verify` for self-signed certificates, or mount CA certificates:
```yaml
volumes:
  - name: ca-certs
    configMap:
      name: kibana-ca
volumeMounts:
  - name: ca-certs
    mountPath: /etc/ssl/certs/kibana-ca.crt
    subPath: ca.crt
```
