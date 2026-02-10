<!-- [DEPRECATED] Use native workflows in .agent/workflows/ instead -->
---
description: How to gather and report bot metrics in a readable format
---

This workflow describes how to pull raw metrics from the bot and format them into a human-readable business intelligence report.

## 1. Quick Report (Recommended)

Use the provided reporting script to get a clean summary of business and technical metrics.

// turbo

```bash
./scripts/report_metrics.sh
```

*Note: If running on the server host, the script defaults to `http://localhost:8083/metrics`.*

## 2. Remote Report

To run the report for a remote server, pass the target URL as an argument:

```bash
./scripts/report_metrics.sh http://your-server-ip:8083/metrics
```

## 3. Raw Metrics (Manual)

If you need the raw Prometheus data for debugging or manual analysis:

```bash
curl http://localhost:8083/metrics | grep vera_
```

## 4. Key Metrics Explained

| Metric Group | Description |
| :--- | :--- |
| **Booking Stats** | Total bookings, returning vs. first-time ratio, and cancellations. |
| **Popularity** | Breakdown of which specific services are favored by patients. |
| **Clinical Depth** | Average length of therapist notes (proxy for clinical engagement). |
| **Technical Health** | Active sessions, DB errors, and OAuth token longevity. |
| **API Latency** | Performance metrics for Google Calendar and Groq integration. |
