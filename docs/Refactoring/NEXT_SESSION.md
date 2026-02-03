# Next Session Plan: Deployment & Monitoring

**Goal**: Deploy the refactored bot to production and verify stability.

---

## 📋 PRIORITIES

### 1. 🚀 Deployment

- [ ] Connect to the Home Server.
- [ ] Pull latest code (`git pull`).
- [ ] Run deployment script (`./scripts/deploy_home_server.sh`).
- [ ] Verify container status (`docker compose ps`).

### 2. 🔍 Post-Deployment Verification

- [ ] Check logs for any startup errors (`docker compose logs -f massage-bot`).
- [ ] Verify functionality:
  - `/start` command.
  - Appointment booking flow.
  - "My Appointments" history.
- [ ] Check Prometheus metrics (`http://SERVER_IP:8083/metrics`).

### 3. 🧹 Final Cleanup (Low Priority)

- [ ] Replace deprecated `io/ioutil` in `googlecalendar/client.go`.
- [ ] Handle ignored errors in Telegram handlers (add error logging).

---

## 🛠️ INSTRUCTIONS

### Step 1: Deploy

1. SSH into server.
2. `cd /opt/vera-bot`
3. `./scripts/deploy_home_server.sh`

### Step 2: Emergency Rollback (If needed)

If deployment fails:

1. `git reset --hard HEAD^` (or previous stable commit)
2. `./scripts/deploy_home_server.sh`
