# Kuerzen Monitoring Stack

This directory contains the complete monitoring setup for the application using:

- **Grafana Alloy**: Unified observability agent for collecting logs and metrics
- **Prometheus**: Time-series metrics database
- **Loki**: Log aggregation system  
- **Grafana**: Visualization and dashboards

## Components

### Grafana Alloy

- **Config**: `alloy/config.alloy`
- **Purpose**: Collects logs and metrics from services of Docker containers
  - Forwards logs to Loki and metrics to Prometheus
- **UI**: http://localhost:12345

### Prometheus  

- **Location**: `prometheus/prometheus.yml`
- **Purpose**: Stores time-series metrics data
- **UI**: http://localhost:9090

### Loki

- **Location**: `loki/loki-config.yml`
- **Purpose**: Stores and indexes log data

### Grafana

- **Location**: `grafana/`
- **Purpose**: Visualization and dashboards
- **UI**: http://localhost:3030

## Quick Start

1.**Add monitoring environment variables to your `.monitoring.env` file**:

In the root directory of the project, create or edit the `.monitoring.env` file with the following content:

```env
# Grafana Configuration
GF_SECURITY_ADMIN_USER=admin
GF_SECURITY_ADMIN_PASSWORD=passwd
GF_USERS_ALLOW_SIGN_UP=false

# Application Services to Monitor (for service discovery)
APP_ENV=development
SHORTENER_PORT=3000
REDIRECTOR_PORT=3001
ANALYTICS_PORT=3002
```

2.**Start the complete stack**:

In the root directory of the project, run:

```bash
docker compose -f compose.monitoring.yml up -d
```

3.**Access dashboards**:

- Grafana: http://localhost:3030
- Prometheus: http://localhost:9090
- Alloy: http://localhost:12345

4.**View logs and metrics**:

- View logs: go to Grafana UI -> Sidebar -> Drilldown -> Logs
- View metrics: go to Grafana UI -> Sidebar -> Drilldown -> Metrics

## Metrics Endpoints

Each Go service exposes:

- `/health` - Health check (JSON response)
- `/metrics` - Prometheus metrics (text format)
