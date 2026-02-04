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

        async function cancelAppointment(apptId, btn) {
            const tg = window.Telegram.WebApp;
            const url = new URL(window.location.href);
            const id = url.searchParams.get('id');
            const token = url.searchParams.get('token');

            // Add loading state to button
            if (btn) {
                btn.classList.add('loading');
                btn.textContent = 'Отмена...';
            }

            try {
                const resp = await fetch("/cancel?id=" + id + "&token=" + token + "&apptId=" + apptId, { 
                    method: "POST",
                    headers: {
                        "ngrok-skip-browser-warning": "true"
                    }
                });
                const result = await resp.json();
                
                if (result.status === "ok") {
                    location.reload();
                } else {
                    // Remove loading state on error
                    if (btn) {
                        btn.classList.remove('loading');
                        btn.textContent = 'Отменить';
                    }
                    tg.showAlert("Ошибка: " + (result.error || "Не удалось отменить запись. Возможно, до приема менее 72 часов."));
                }
            } catch (e) {
                // Remove loading state on network error
                if (btn) {
                    btn.classList.remove('loading');
                    btn.textContent = 'Отменить';
                }
                tg.showAlert("Ошибка сети при отмене записи.");
            }
        }

        function updateCountdown() {
            const nextUnix = {{.NextApptUnix}};
            if (nextUnix === 0) return;

            const now = Math.floor(Date.now() / 1000);
            const diff = nextUnix - now;
            const el = document.getElementById('countdown');
            if (!el) return;

            if (diff <= 0) {
                el.innerText = "Прием начинается...";
                return;
            }

            const days = Math.floor(diff / 86400);
            const hours = Math.floor((diff % 86400) / 3600);
            const mins = Math.floor((diff % 3600) / 60);

            let str = "";
            if (days > 0) str += days + "д ";
            if (hours > 0 || days > 0) str += hours + "ч ";
            str += mins + "м";
            
            el.innerText = "До приема: " + str;
        }

        window.addEventListener('DOMContentLoaded', () => {
            const tg = window.Telegram.WebApp;
            if (tg && tg.expand) { tg.expand(); tg.ready(); tg.setHeaderColor('#ffffff'); }
            
            updateCountdown();
            setInterval(updateCountdown, 60000);
        });
    </script>
    <style>
        :root {
            --accent: #2563eb; --accent-soft: #eff6ff; --bg-page: #f8fafc; --bg-card: #ffffff;
            --text-main: #0f172a; --text-muted: #64748b; --border: #e2e8f0; --glass: rgba(255, 255, 255, 0.85);
            --danger: #ef4444; --danger-soft: #fef2f2; --success: #22c55e;
        }
        
        /* Dark Mode Support */
        @media (prefers-color-scheme: dark) {
            :root {
                --accent: #3b82f6; --accent-soft: #1e3a5f; --bg-page: #0f172a; --bg-card: #1e293b;
                --text-main: #f1f5f9; --text-muted: #94a3b8; --border: #334155; --glass: rgba(30, 41, 59, 0.85);
                --danger: #f87171; --danger-soft: #3f1e1e; --success: #4ade80;
            }
            .notes-content { color: #cbd5e1; }
        }
        
        * { box-sizing: border-box; -webkit-font-smoothing: antialiased; }
        body { background-color: var(--bg-page); font-family: 'Inter', system-ui, sans-serif; margin: 0; padding: 0; color: var(--text-main); line-height: 1.6; overflow-x: hidden; }
        
        /* Section Animations */
        @keyframes fadeSlideIn {
            from { opacity: 0; transform: translateY(12px); }
            to { opacity: 1; transform: translateY(0); }
        }
        section { 
            background: var(--bg-card); border-radius: 16px; padding: 24px; border: 1px solid var(--border); 
            margin-bottom: 24px; box-shadow: 0 1px 3px rgba(0,0,0,0.02);
            animation: fadeSlideIn 0.4s ease-out backwards;
        }
        section:nth-child(1) { animation-delay: 0.05s; }
        section:nth-child(2) { animation-delay: 0.1s; }
        section:nth-child(3) { animation-delay: 0.15s; }
        section:nth-child(4) { animation-delay: 0.2s; }
        section:nth-child(5) { animation-delay: 0.25s; }
        
        /* Future Appointments - Prominent */
        section.upcoming { border-left: 3px solid var(--accent); }
        
        /* Past/History - Subtle */
        section.history { opacity: 0.9; }
        
        .premium-header { background: var(--bg-card); padding: 32px 24px; border-bottom: 1px solid var(--border); position: sticky; top: 0; z-index: 50; backdrop-filter: blur(12px); background: var(--glass); animation: fadeSlideIn 0.3s ease-out; }
        .header-content { max-width: 800px; margin: 0 auto; }
        .badge { display: inline-block; padding: 4px 10px; background: var(--accent-soft); color: var(--accent); border-radius: 20px; font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 12px; }
        h1 { font-family: 'Outfit', sans-serif; font-size: 28px; font-weight: 700; margin: 0 0 8px 0; color: var(--text-main); letter-spacing: -0.02em; }
        .patient-meta { font-size: 13px; color: var(--text-muted); display: flex; align-items: center; gap: 8px; }
        .stat-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-top: 24px; }
        .stat-card { background: var(--bg-page); padding: 16px 8px; border-radius: 12px; border: 1px solid var(--border); display: flex; flex-direction: column; justify-content: space-between; align-items: center; min-height: 100px; text-align: center; transition: transform 0.2s ease, box-shadow 0.2s ease; }
        .stat-card:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
        .stat-val { font-size: 15px; font-weight: 800; color: var(--accent); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; width: 100%; margin: auto 0; }
        .stat-val-large { font-size: 20px; }
        .stat-desc { font-size: 10px; text-transform: uppercase; font-weight: 600; color: var(--text-muted); padding-top: 8px; }
        .main-container { max-width: 800px; margin: 0 auto; padding: 24px; }
        h2 { font-family: 'Outfit', sans-serif; font-size: 14px; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-muted); margin: 0 0 16px 0; display: flex; align-items: center; gap: 8px; }
        .notes-content { font-size: 16px; color: #334155; white-space: pre-wrap; line-height: 1.6; }
        .notes-content h1, .notes-content h2, .notes-content h3 { color: var(--text-main); margin-top: 24px; margin-bottom: 12px; font-family: 'Outfit', sans-serif; }
        .notes-content h1 { font-size: 20px; }
        .notes-content h2 { font-size: 18px; }
        .notes-content h3 { font-size: 16px; }
        .doc-list { display: flex; flex-direction: column; gap: 8px; }
        .doc-item { display: flex; justify-content: space-between; align-items: center; padding: 14px; background: var(--bg-page); border-radius: 12px; text-decoration: none; color: var(--text-main); font-size: 14px; font-weight: 500; border: 1px solid transparent; transition: all 0.2s ease; }
        .doc-item:hover { transform: translateY(-1px); border-color: var(--accent); }
        .doc-info { display: flex; flex-direction: column; }
        .doc-stat { font-size: 12px; color: var(--text-muted); font-weight: 400; margin-top: 2px; }
        .doc-latest { font-size: 11px; color: var(--text-muted); font-weight: 400; text-align: right; }
        .footer { text-align: center; padding: 32px 24px 64px; color: var(--text-muted); font-size: 12px; font-weight: 500; }
        
        /* Cancel Button with Loading State */
        .btn-cancel { 
            padding: 6px 12px; background: var(--danger-soft); color: var(--danger); 
            border: none; border-radius: 8px; font-size: 12px; font-weight: 700; 
            cursor: pointer; transition: all 0.2s ease; min-width: 80px;
        }
        .btn-cancel:hover { background: #fee2e2; transform: scale(1.02); }
        .btn-cancel:active { transform: scale(0.98); }
        .btn-cancel:focus-visible { outline: 2px solid var(--danger); outline-offset: 2px; }
        .btn-cancel.loading { 
            opacity: 0.7; pointer-events: none; 
            background: var(--border); color: var(--text-muted);
        }
        .btn-cancel.loading::after { content: " ⏳"; }
        
        .appt-item { display: flex; justify-content: space-between; align-items: center; padding: 16px 0; border-bottom: 1px solid var(--border); }
        .appt-item:last-child { border-bottom: none; }
        .countdown-banner { background: var(--accent); color: white; padding: 6px 12px; border-radius: 8px; font-size: 12px; font-weight: 700; margin-top: 12px; display: inline-block; animation: pulse 2s infinite; }
        @keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.8; } }
        .contact-vera { font-size: 11px; color: var(--accent); text-decoration: none; font-weight: 600; display: inline-flex; align-items: center; gap: 4px; transition: opacity 0.2s; }
        .contact-vera:hover { opacity: 0.8; }
        
        /* Empty State */
        .empty-state { text-align: center; padding: 32px 20px; color: var(--text-muted); font-size: 14px; }
        .empty-state-icon { font-size: 32px; margin-bottom: 8px; opacity: 0.5; }

        /* Mobile Optimization */
        @media (max-width: 480px) {
            .premium-header { padding: 24px 16px; }
            h1 { font-size: 24px; }
            .stat-grid { grid-template-columns: 1fr; gap: 8px; }
            .stat-card { flex-direction: row; min-height: 48px; padding: 12px 16px; border-radius: 10px; }
            .stat-card:hover { transform: none; box-shadow: none; }
            .stat-val { text-align: right; width: auto; margin: 0; font-size: 13px; font-weight: 700; }
            .stat-val-large { font-size: 16px; }
            .stat-desc { padding-top: 0; font-size: 9px; order: -1; text-align: left; }
            .main-container { padding: 12px; }
            section { padding: 16px; border-radius: 12px; margin-bottom: 16px; }
            .notes-content { font-size: 14px; }
        }
    </style>
</head>
<body>
    <header class="premium-header">
        <div class="header-content">
            <span class="badge">КАРТА ПАЦИЕНТА</span>
            <h1>{{.Name}}</h1>
            <div class="patient-meta">
                <span>ID: {{.TelegramID}}</span>
            </div>
            <div class="stat-grid">
                <div class="stat-card">
                    <div class="stat-desc">Посещений</div>
                    <div class="stat-val stat-val-large">{{.TotalVisits}}</div>
                </div>
                <div class="stat-card">
                    <div class="stat-desc">Услуга</div>
                    <div class="stat-val">{{.CurrentService}}</div>
                </div>
                <div class="stat-card">
                    <div class="stat-desc">Обновлено</div>
                    <div class="stat-val">{{.LastVisit}}</div>
                </div>
            </div>
        </div>
    </header>
    <main class="main-container">
        <!-- UPCOMING APPOINTMENTS -->
        {{if .FutureAppointments}}
        <section class="upcoming">
            <h2>Будущие записи</h2>
            <div class="appt-list">
                {{range .FutureAppointments}}
                <div class="appt-item">
                    <div>
                        <div style="font-weight: 700; font-size: 15px;">{{.Date}}</div>
                        <div style="font-size: 12px; color: var(--text-muted);">{{.Service}}</div>
                    </div>
                    <div>
                        {{if .CanCancel}}
                            <button class="btn-cancel" onclick="cancelAppointment('{{.ID}}', this)">Отменить</button>
                        {{else}}
                            <a href="https://t.me/VeraFethiye" class="contact-vera">💬 Написать Вере</a>
                        {{end}}
                    </div>
                </div>
                {{end}}
                
                {{if .NextApptUnix}}
                <div id="countdown" class="countdown-banner">⏳ Загрузка...</div>
                {{end}}
                
                <p style="font-size: 11px; color: var(--text-muted); margin-top: 12px;">
                    ⚠️ Отмена возможна за 72 часа до приема.
                </p>
            </div>
        </section>
        {{end}}

        <section>
            <h2>История болезни</h2>
            <div class="notes-content">{{.TherapistNotes}}</div>
        </section>

        {{if .RecentVisits}}
        <section class="history">
            <h2>История посещений</h2>
            <div class="appt-list">
                {{range .RecentVisits}}
                <div class="appt-item">
                    <div style="font-weight: 600; font-size: 14px;">{{.Date}}</div>
                    <div style="font-size: 13px; color: var(--text-muted);">{{.Service}}</div>
                </div>
                {{end}}
            </div>
        </section>
        {{end}}

        {{if .VoiceTranscripts}}
        <section>
            <h2>Расшифровки Консультаций</h2>
            <div class="notes-content" style="font-style: italic; color: #64748b;">{{.VoiceTranscripts}}</div>
        </section>
        {{end}}

        <section>
            <h2>Документы и Снимки</h2>
            <div class="doc-list">
                {{range .DocGroups}}
                <div class="doc-item">
                    <div class="doc-info">
                        <div style="font-weight: 600;">{{.Name}}</div>
                        <div class="doc-stat">Количество: {{.Count}}</div>
                    </div>
                    <div class="doc-latest">
                        <div>Последний:</div>
                        <div>{{.Latest}}</div>
                    </div>
                </div>
                {{else}}
                <div style="text-align: center; padding: 20px; color: var(--text-muted); font-size: 14px;">Документов пока нет.</div>
                {{end}}
            </div>
        </section>
        <footer class="footer">Vera Massage Bot {{.BotVersion}}<br>Professional Medical Record Hub</footer>
    </main>
</body>
</html>
`
