# 📊 Metrics Reference (Memo)

Overview of the clinical and technical metrics currently instrumented in the bot. All metrics are prefixed with `vera_`.

## 📈 Business & Patient Behavior

| Metric | Type | Labels | Purpose |
| :--- | :--- | :--- | :--- |
| `vera_bookings_total` | Counter | `service` | Total volume of confirmed appointments. |
| `vera_appointment_type_total`| Counter | `type` | Tracks **Loyalty**: `first_visit` vs `returning`. |
| `vera_booking_lead_time_days` | Histogram | - | Measures **Behavior**: How many days in advance users book. |
| `vera_service_bookings_total`| Counter | `service_name` | Measures **Service Popularity**. |
| `vera_clinical_note_length_chars`| Gauge | - | Measures **Engagement Depth** (avg length of therapist notes). |
| `vera_booking_creation_hour_total`| Counter | `hour` | Identifies peak booking activity hours. |
| `vera_cancellations_total` | Counter | `service_name` | Monitor dropout rates/cancellations. |

## 💻 Technical Health

| Metric | Type | Labels | Purpose |
| :--- | :--- | :--- | :--- |
| `vera_api_requests_total` | Counter | `provider`, `operation`, `status` | Reliability tracking for Google & Groq APIs. |
| `vera_api_latency_seconds` | Histogram | `provider`, `operation` | Performance monitoring for external dependencies. |
| `vera_db_errors_total` | Counter | `operation` | Database stability tracking. |
| `vera_active_sessions` | Gauge | - | Real-time concurrent bot users. |
| `vera_token_expiry_days` | Gauge | - | Warning system for Google OAuth expiration. |

---

## 🛠️ Access & Harvesting

Metrics are exposed in Prometheus format on the internal health server:

- **Port**: `8083` (Changed from 8081 to avoid conflict with cadvisor)
- **Path**: `/metrics`
- **Dashboard**: Use the `scripts/report_metrics.sh` for a human-friendly view.
