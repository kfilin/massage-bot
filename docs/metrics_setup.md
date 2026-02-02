# Monitoring Setup

This guide explains how to integrate **Massage Bot** with your existing monitoring stack (Prometheus + Grafana).

The configuration files are located in `deploy/monitoring/`.

## Prerequisites

- A running Prometheus instance.
- A running Grafana instance.
- The Massage Bot running and exposing metrics on port `8083`.

## 1. Configure Prometheus

Append the contents of `deploy/monitoring/prometheus_job.yml` to your main `prometheus.yml` configuration file (usually found in `/etc/prometheus/` or your tracking stack volume).

```yaml
scrape_configs:
  # ... your existing jobs ...

  - job_name: 'massage-bot'
    scrape_interval: 15s
    metrics_path: '/metrics'
    static_configs:
      - targets: ['172.17.0.1:8083'] # REPLACE with your bot's IP/Host
```

**Networking Note:**

- If Prometheus is running in a Docker container on the same host, use the host's IP address (e.g., `172.17.0.1` for the docker0 bridge) or `host.docker.internal` config depending on your OS.
- If they share a network, use the container name `massage-bot:8083`.

## 2. Import Grafana Dashboard

1. Log in to your Grafana instance.
2. Navigate to **Dashboards** -> **Import**.
3. Upload the `deploy/monitoring/grafana_dashboard.json` file.
4. Select your Prometheus data source.
5. Click **Import**.

## 3. Verify

Check the "Massage Bot Dashboard" in Grafana. You should see metrics for:

- Total Bookings
- Active Sessions
- Service Popularity
- DB Errors
