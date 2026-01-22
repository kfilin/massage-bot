package storage

const medicalRecordTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no, viewport-fit=cover">
    <title>Медицинская карта</title>
    <script src="https://telegram.org/js/telegram-web-app.js"></script>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&family=Outfit:wght@400;600;700&display=swap" rel="stylesheet">
    <script>
        // Auth Self-Healing Logic
        (function() {
            const tg = window.Telegram.WebApp;
            const url = new URL(window.location.href);
            if (tg.initData && !url.searchParams.get('initData')) {
                url.searchParams.set('initData', tg.initData);
                window.location.replace(url.toString());
            }
        })();

        async function cancelAppointment(apptId) {
            const tg = window.Telegram.WebApp;
            const url = new URL(window.location.href);
            const id = url.searchParams.get('id');
            const token = url.searchParams.get('token');

            if (!confirm("Вы уверены, что хотите отменить запись?")) return;

            try {
                const resp = await fetch("/cancel?id=" + id + "&token=" + token + "&apptId=" + apptId, { method: "POST" });
                const result = await resp.json();
                
                if (result.status === "ok") {
                    location.reload();
                } else {
                    tg.showAlert("Ошибка: " + (result.error || "Не удалось отменить запись. Возможно, до приема менее 72 часов."));
                }
            } catch (e) {
                tg.showAlert("Ошибка сети при отмене записи.");
            }
        }

        window.addEventListener('DOMContentLoaded', () => {
            const tg = window.Telegram.WebApp;
            if (tg && tg.expand) { tg.expand(); tg.ready(); tg.setHeaderColor('#ffffff'); }
        });
    </script>
    <style>
        :root {
            --accent: #2563eb; --accent-soft: #eff6ff; --bg-page: #f8fafc; --bg-card: #ffffff;
            --text-main: #0f172a; --text-muted: #64748b; --border: #e2e8f0; --glass: rgba(255, 255, 255, 0.85);
            --danger: #ef4444; --danger-soft: #fef2f2;
        }
        * { box-sizing: border-box; -webkit-font-smoothing: antialiased; }
        body { background-color: var(--bg-page); font-family: 'Inter', system-ui, sans-serif; margin: 0; padding: 0; color: var(--text-main); line-height: 1.6; overflow-x: hidden; }
        .premium-header { background: var(--bg-card); padding: 32px 24px; border-bottom: 1px solid var(--border); position: sticky; top: 0; z-index: 50; backdrop-filter: blur(12px); background: var(--glass); }
        .header-content { max-width: 800px; margin: 0 auto; }
        .badge { display: inline-block; padding: 4px 10px; background: var(--accent-soft); color: var(--accent); border-radius: 20px; font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 12px; }
        h1 { font-family: 'Outfit', sans-serif; font-size: 28px; font-weight: 700; margin: 0 0 8px 0; color: var(--text-main); letter-spacing: -0.02em; }
        .patient-meta { font-size: 13px; color: var(--text-muted); display: flex; align-items: center; gap: 12px; }
        .stat-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-top: 24px; }
        .stat-card { background: var(--bg-page); padding: 12px; border-radius: 12px; border: 1px solid var(--border); text-align: center; }
        .stat-val { font-size: 18px; font-weight: 800; color: var(--accent); }
        .stat-desc { font-size: 10px; text-transform: uppercase; font-weight: 600; color: var(--text-muted); }
        .main-container { max-width: 800px; margin: 0 auto; padding: 24px; }
        section { background: var(--bg-card); border-radius: 16px; padding: 24px; border: 1px solid var(--border); margin-bottom: 24px; box-shadow: 0 1px 3px rgba(0,0,0,0.02); }
        h2 { font-family: 'Outfit', sans-serif; font-size: 14px; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-muted); margin: 0 0 16px 0; display: flex; align-items: center; gap: 8px; }
        h2::before { content: ""; width: 4px; height: 14px; background: var(--accent); border-radius: 4px; }
        .notes-content { font-size: 16px; color: #334155; white-space: pre-wrap; line-height: 1.8; }
        .doc-list { display: flex; flex-direction: column; gap: 8px; }
        .doc-item { display: flex; align-items: center; gap: 12px; padding: 14px; background: var(--bg-page); border-radius: 12px; text-decoration: none; color: var(--text-main); font-size: 14px; font-weight: 500; border: 1px solid transparent; transition: all 0.2s; }
        .footer { text-align: center; padding: 32px 24px 64px; color: var(--text-muted); font-size: 12px; font-weight: 500; }
        .btn-cancel { padding: 6px 12px; background: var(--danger-soft); color: var(--danger); border: none; border-radius: 8px; font-size: 12px; font-weight: 700; cursor: pointer; }
        .appt-item { display: flex; justify-content: space-between; align-items: center; padding: 12px 0; border-bottom: 1px solid var(--border); }
        .appt-item:last-child { border-bottom: none; }
    </style>
</head>
<body>
    <header class="premium-header">
        <div class="header-content">
            <span class="badge">Клиническая Карта</span>
            <h1>{{.Name}}</h1>
            <div class="patient-meta">
                <span>ID: {{.TelegramID}}</span>
                <span>•</span>
                <span>Visits: {{.TotalVisits}}</span>
            </div>
            <div class="stat-grid">
                <div class="stat-card">
                    <div class="stat-val">{{.TotalVisits}}</div>
                    <div class="stat-desc">Посещений</div>
                </div>
                <div class="stat-card">
                    <div class="stat-val" style="font-size: 12px;">{{.CurrentService}}</div>
                    <div class="stat-desc">Услуга</div>
                </div>
                <div class="stat-card">
                    <div class="stat-val" style="font-size: 12px;">{{.LastVisit}}</div>
                    <div class="stat-desc">Обновлено</div>
                </div>
            </div>
        </div>
    </header>
    <main class="main-container">
        <!-- UPCOMING APPOINTMENTS -->
        {{if .ShowLastVisitLink}}
        <section>
            <h2>Будущие записи</h2>
            <div class="appt-list">
                <div class="appt-item">
                    <div>
                        <div style="font-weight: 700; font-size: 15px;">{{.LastVisit}}</div>
                        <div style="font-size: 12px; color: var(--text-muted);">{{.CurrentService}}</div>
                    </div>
                </div>
                <p style="font-size: 11px; color: var(--text-muted); margin-top: 12px;">
                    ⚠️ Отмена возможна за 72 часа до приема. Для отмены свяжитесь с терапевтом.
                </p>
            </div>
        </section>
        {{end}}

        <section>
            <h2>История Болезни</h2>
            <div class="notes-content">{{.TherapistNotes}}</div>
        </section>

        {{if .VoiceTranscripts}}
        <section>
            <h2>Расшифровки Консультаций</h2>
            <div class="notes-content" style="font-style: italic; color: #64748b;">{{.VoiceTranscripts}}</div>
        </section>
        {{end}}

        <section>
            <h2>Документы и Снимки</h2>
            <div class="doc-list">
                {{range .Documents}}
                <div class="doc-item">
                    <span class="doc-icon">📁</span>
                    <span>{{.Name}}</span>
                </div>
                {{else}}
                <div style="text-align: center; padding: 20px; color: var(--text-muted); font-size: 14px;">Документов пока нет.</div>
                {{end}}
            </div>
        </section>
        <footer class="footer">Vera Massage Bot v4.1.0<br>Professional Medical Record Hub</footer>
    </main>
</body>
</html>
`
