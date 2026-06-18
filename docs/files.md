# 📂 Project File Structure

---

## 📁 Root

| Entry | Purpose |
| :--- | :--- |
| `cmd/bot/` | Application entry point, health server, web app routing |
| `internal/` | Core domain logic, services, storage, delivery, adapters |
| `deploy/` | Docker Compose, Caddyfiles, K8s manifests, Grafana dashboard |
| `scripts/` | Deployment, backup, and metric scripts |
| `docs/` | Technical and user documentation |
| `data/` | Clinical records (`.md` files), backups, patient folders |
| `.pi/skills/` | AI agent skills (startup, handoff, hydrate-harness) |
| `global-skills/` | Project-agnostic methodology library (TDD, debugging, etc.) |

### Config & Build

| File | Purpose |
| :--- | :--- |
| `go.mod` / `go.sum` | Go module dependencies |
| `Makefile` | Build, test, lint, coverage shortcuts |
| `Dockerfile` | Container build instructions |
| `docker-compose.yml` | Main orchestration (production) |
| `.env.example` | Environment variable template |
| `.gitlab-ci.yml` | GitLab CI/CD pipeline |

---

## 💻 Source Code (`internal/`)

| Package | Purpose |
| :--- | :--- |
| `internal/domain/` | Core entities: Patient, Appointment, Slot |
| `internal/services/appointment/` | Booking engine: slot search, conflict detection |
| `internal/services/reminder/` | Reminder worker: 72h/24h ticker lifecycle |
| `internal/storage/` | PostgreSQL repository, session storage, file mirroring, migration |
| `internal/delivery/telegram/` | Bot handlers, callback routing, middleware, keyboards |
| `internal/delivery/web/` | TWA HTTP handlers: medical card, search, drafts, cancel, transcribe |
| `internal/adapters/googlecalendar/` | Google Calendar Free/Busy API adapter |
| `internal/adapters/groq/` | Groq Whisper transcription adapter |
| `internal/ports/` | Interface boundaries (BotAPI, Repository, AppointmentService) |
| `internal/presentation/` | HTML templates (TWA), Telegram message formatters |
| `internal/config/` | Environment variable parsing and configuration |
| `internal/logging/` | Structured logger (zerolog-style) |
| `internal/monitoring/` | Prometheus metrics collectors |
| `internal/version/` | Build version constant |

---

## 🚀 Deploy & Infrastructure (`deploy/`)

| File / Directory | Purpose |
| :--- | :--- |
| `docker-compose.yml` | Main production orchestration |
| `deploy/docker-compose.prod.yml` | Production-specific overrides (ports, env) |
| `deploy/docker-compose.dev.yml` | Development overrides (Caddy) |
| `deploy/Caddyfile` | Production reverse proxy config (HTTPS) |
| `deploy/Caddyfile.test` | Test environment proxy |
| `deploy/Caddyfile.dev` | Local development proxy |
| `deploy/k8s/` | Kubernetes manifests (deployment, service, configmap, secrets) |
| `deploy/monitoring/grafana_dashboard.json` | Grafana dashboard export |
| `deploy/monitoring/prometheus_job.yml` | Prometheus scrape config snippet |

---

## 🔧 Scripts (`scripts/`)

| File | Purpose |
| :--- | :--- |
| `scripts/deploy.sh` | **Primary deploy wrapper**: port-collision pre-flight, compose-based deploy for test/prod |
| `scripts/deploy_home_server.sh` | Legacy deploy (kept for GitLab CI compatibility) |
| `scripts/deploy_test_server.sh` | Legacy test deploy (kept for GitLab CI compatibility) |
| `scripts/verify_backup.sh` | Backup integrity verification (ZIP, entries, JSON) |
| `scripts/report_metrics.sh` | Structured console metrics view |

---

## 📜 Documentation

| File | Audience | Purpose |
| :--- | :--- | :--- |
| `README.md` | Everyone | Project overview, features, quick start |
| `AGENTS.md` | AI agents | Project rules, guardrails, DOX framework |
| `AGENT_USER_MANUAL.md` | Developers | Agent collaboration conventions |
| `USER_GUIDE.md` | Patients | Booking, cancel, medical card (EN) |
| `USER_GUIDE_RU.md` | Patients | Booking, cancel, medical card (RU) |
| `DEVELOPER.md` | Developers | Architecture, testing, deployment |
| `CHANGELOG.md` | Everyone | Release history |
| `BACKLOG.md` | AI agents | Active tasks, bugs, session journal |
| `docs/API.md` | Developers | HTTP endpoints, Prometheus metrics |
| `docs/VERA_GUIDE_RU.md` | Therapist | Clinical workflow, admin commands (RU) |
| `docs/CI_CD_Pipeline.md` | Developers | Deployment pipeline, twin strategy |
| `docs/metrics_setup.md` | Developers | Prometheus/Grafana integration guide |
| `docs/ProdArchitecture.md` | Developers | Current vs Enterprise architecture gap analysis |
| `docs/backlog_design.md` | Developers | UI/UX design ideas (future consideration) |
| `metrics.md` | Developers | Full Prometheus metrics reference |
| `data/README.md` | Developers | Patient data directory structure |

---

## 📦 Data & Storage

| Directory | Purpose |
| :--- | :--- |
| `data/patients/` | Mirrored clinical Markdown files (one folder per patient) |
| `data/backups/` | ZIP archives staging area for Telegram delivery |
| `logs/` | Application access logs |

---

## 🗄️ Archive

| Directory | Contents |
| :--- | :--- |
| `ARCHIVE/HANDOFF/` | Historical session handoffs |
| `ARCHIVE/LAST_SESSION/` | Historical last-session notes |
| `ARCHIVE/Refactoring/` | Obsolete refactoring documentation |
| `ARCHIVE/CHANGELOGS/` | Legacy changelogs |
| `AGENTIC_OS_TEMPLATE/` | Agentic OS template files (reference only) |

---

*Last updated: 2026-06-18.*
