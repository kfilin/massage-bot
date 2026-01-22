# Handoff: Clinical Massage Bot (v4.1.0)

## 🎯 Status: MISSION OPERATIONAL

The project has been successfully restored to a stable production state. All "PDF Experimental" code has been purged.

## 🔑 Key Architectural Pins

- **Zero PDF Policy**: PDF generation is dead. We use a **Live TWA Card** with a clinical white theme.
- **Clinical Storage 2.0**: The system mirrors PostgreSQL to `.md` files. This is the **primary** way the therapist interacts with data via WebDAV.
- **WebDAV**: Address: `https://[DOMAIN]/webdav/`. It follows the basic auth from `.env` (`WEBAPP_SECRET`).
- **Reminders**: The worker runs every 15 mins and checks for visits in the `(1h 45m - 2h 15m)` window.

## ⚠️ Critical Files

- `internal/storage/postgres_repository.go`: Contains the sync/mirroring logic.
- `internal/delivery/telegram/reminders.go`: Contains the 2h notification logic.
- `internal/storage/record_template.go`: The TWA HTML/CSS (PDF-free).

## 🚀 Future Directives

1. **Always use Go 1.24** (Alpine).
2. **Never re-introduce local PDF generation libraries** (gofpdf, etc.).
3. **Istanbul Timezone** must be preserved in all Docker/App contexts.

---
*Ready for the next session.*
