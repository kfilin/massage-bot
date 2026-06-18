# Monitoring Setup

This directory contains configuration files to integrate **Massage Bot** with Prometheus + Grafana.

## Files

| File | Purpose |
|------|---------|
| `grafana_dashboard.json` | Grafana dashboard export (15+ panels) |
| `prometheus_job.yml` | Prometheus scrape config snippet |
| `README.md` | This file |

## Quick Setup

1. Add `prometheus_job.yml` content to your Prometheus config.
2. Import `grafana_dashboard.json` into Grafana.
3. Verify metrics at `http://<bot-host>:8083/metrics`.

See [Metrics Setup Guide](../docs/metrics_setup.md) for detailed instructions.

---
*Last updated: 2026-06-18.*