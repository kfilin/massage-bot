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

        async function cancelAppointment(event, apptId, btn) {
            if (event) {
                event.preventDefault();
                event.stopPropagation();
            }
            const tg = window.Telegram.WebApp;

            // Add loading state to button
            if (btn) {
                btn.classList.add('loading');
                btn.textContent = 'Отмена...';
            }

            try {
                // Use Telegram's native initData for auth (never expires)
                const resp = await fetch("/cancel", { 
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                        "ngrok-skip-browser-warning": "true"
                    },
                    body: JSON.stringify({
                        initData: tg.initData,
                        apptId: apptId
                    })
                });
                const result = await resp.json();
                
                if (result.status === "ok") {
                    // Fix: Avoid location.reload() to prevent TWA redirect issues.
                    // Instead, update the DOM directly.
                    tg.showAlert("✅ Запись успешно отменена");
                    
                    if (btn) {
                        // Find the row and visualy remove it
                        const row = btn.closest('.appt-item');
                        if (row) {
                            row.style.transition = 'all 0.5s ease';
                            row.style.opacity = '0';
                            row.style.transform = 'translateX(20px)';
                            setTimeout(() => {
                                row.remove();
                                // Check if list is empty? Optional.
                            }, 500);
                        }
                        
                        // Also try to hide the Next Appointment card if it matches (optional heuristics)
                        // Or just advise user to reload if they want fresh stats
                    }
                } else {
                    // Remove loading state on error
                    if (btn) {
                        btn.classList.remove('loading');
                        btn.textContent = 'Отменить';
                    }
                    tg.showAlert("Ошибка: " + (result.error || "Не удалось отменить запись."));
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

        function toggleSection(header) {
            const section = header.closest('section');
            const isCollapsed = section.classList.toggle('collapsed');
            header.setAttribute('aria-expanded', !isCollapsed);
        }

        // Keyboard support for collapsibles
        function handleKey(e, header) {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                toggleSection(header);
            }
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
        

        /* Accessibility */
        :focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
        button:focus-visible, a:focus-visible, .collapsible-header:focus-visible { outline: 2px solid var(--accent); outline-offset: 4px; }
        
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
        .stat-card { background: var(--bg-page); padding: 16px 12px; border-radius: 12px; border: 1px solid var(--border); display: flex; flex-direction: column; justify-content: space-between; align-items: center; min-height: 110px; text-align: center; transition: all 0.2s ease; }
        .stat-card:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.08); border-color: var(--accent); }
        .stat-val { font-size: 15px; font-weight: 800; color: var(--text-main); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; width: 100%; margin: auto 0; }
        .stat-val-large { font-size: 22px; color: var(--accent); }
        .stat-desc { font-size: 10px; text-transform: uppercase; font-weight: 700; color: var(--text-muted); letter-spacing: 0.05em; }
        .stat-sub { font-size: 10px; color: var(--text-muted); margin-top: 4px; font-weight: 500; }
        .next-appt { border: 1.5px solid var(--accent); background: var(--accent-soft); }
        .next-appt .stat-val { color: var(--accent); }
        
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

        /* Primary Action Buttons */
        .btn-primary { 
            display: inline-flex; align-items: center; justify-content: center; gap: 6px;
            padding: 8px 14px; background: var(--accent); color: white; 
            border: none; border-radius: 10px; font-size: 12px; font-weight: 700; 
            text-decoration: none; cursor: pointer; transition: all 0.2s ease;
            box-shadow: 0 4px 12px rgba(51, 144, 236, 0.2);
        }
        .btn-primary:active { transform: scale(0.96); }
        .btn-admin { background: #10b981; box-shadow: 0 4px 12px rgba(16, 185, 129, 0.2); }
        
        /* Empty State */
        .empty-state { text-align: center; padding: 40px 24px; color: var(--text-muted); background: var(--bg-page); border-radius: 12px; border: 1px dashed var(--border); }
        .empty-state-icon { font-size: 40px; margin-bottom: 12px; display: block; filter: grayscale(1); opacity: 0.6; }
        .empty-state-text { font-size: 14px; font-weight: 500; }
        .empty-state-sub { font-size: 12px; margin-top: 4px; opacity: 0.8; }
 
        /* Collapsible Sections */
        .collapsible-header { cursor: pointer; display: flex; justify-content: space-between; align-items: center; }
        .collapsible-header::after { content: '▼'; font-size: 10px; color: var(--text-muted); transition: transform 0.3s ease; }
        section.collapsed .collapsible-header::after { transform: rotate(-90deg); }
        .collapsible-content { max-height: 1000px; overflow: hidden; transition: max-height 0.3s ease, opacity 0.3s ease; opacity: 1; }
        section.collapsed .collapsible-content { max-height: 0; opacity: 0; }
 
        /* Mobile Optimization */
        @media (max-width: 480px) {
            .premium-header { padding: 24px 16px; }
            h1 { font-size: 24px; }
            .stat-grid { grid-template-columns: repeat(2, 1fr); gap: 10px; }
            .stat-card { padding: 14px 10px; min-height: 90px; }
            .stat-card.next-appt { grid-column: span 2; min-height: 80px; flex-direction: row; padding: 12px 16px; justify-content: space-between; text-align: left; }
            .stat-card.next-appt .stat-desc { order: -1; }
            .stat-card.next-appt .stat-val { width: auto; margin: 0; text-align: right; font-size: 14px; }
            .stat-card.next-appt .stat-sub { display: none; }
            
            .stat-val { font-size: 14px; }
            .stat-val-large { font-size: 20px; }
            .stat-desc { font-size: 9px; }
            .stat-sub { font-size: 9px; }
            
            /* Primary Action Buttons - Mobile */
            .btn-primary { 
                padding: 6px 10px; font-size: 11px;
            }
            .main-container { padding: 16px; }
            section { padding: 16px; border-radius: 12px; margin-bottom: 20px; }
            .notes-content { font-size: 14px; }
        }
        
        /* Primary Action Buttons */
        .btn-primary { 
            display: inline-flex; align-items: center; justify-content: center; gap: 6px;
            padding: 8px 14px; background: var(--accent); color: white; 
            border: none; border-radius: 10px; font-size: 12px; font-weight: 700; 
            text-decoration: none; cursor: pointer; transition: all 0.2s ease;
            box-shadow: 0 4px 12px rgba(51, 144, 236, 0.2);
            font-family: 'Inter', sans-serif;
        }
        .btn-primary:active { transform: scale(0.96); }
        .btn-admin { background: #10b981; box-shadow: 0 4px 12px rgba(16, 185, 129, 0.2); }
        .btn-secondary { background: var(--bg-page); color: var(--text-main); border: 1px solid var(--border); box-shadow: none; }

        /* Media Gallery */
        .doc-files { display: none; padding: 10px; gap: 8px; flex-wrap: wrap; background: #f8fafc; border-bottom-left-radius: 12px; border-bottom-right-radius: 12px; }
        .doc-files.open { display: flex; }
        .media-item { width: 80px; height: 80px; border-radius: 8px; overflow: hidden; border: 1px solid var(--border); position: relative; cursor: pointer; background: white; }
        .media-item img { width: 100%; height: 100%; object-fit: cover; }
        .file-icon-box { display: flex; flex-direction: column; align-items: center; justify-content: center; width: 100%; height: 100%; text-align: center; color: var(--text-main); }
        .file-icon { font-size: 24px; margin-bottom: 4px; }
        .file-date { font-size: 10px; color: var(--text-muted); }
        
        /* Lightbox */
        .lightbox { 
            display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%; 
            background: rgba(0,0,0,0.95); z-index: 2000; justify-content: center; align-items: center; 
            flex-direction: column; opacity: 0; transition: opacity 0.3s ease;
            backdrop-filter: blur(5px);
        }
        .lightbox.visible { display: flex; opacity: 1; }
        .lightbox-content { max-width: 95%; max-height: 80vh; object-fit: contain; border-radius: 8px; box-shadow: 0 4px 12px rgba(0,0,0,0.5); }
        .lightbox-close { 
            position: absolute; top: 20px; right: 20px; color: white; font-size: 30px; 
            cursor: pointer; background: none; border: none; padding: 10px; z-index: 2001;
        }

    </style>
</head>
<body>
    <header class="premium-header">
        <div class="header-content">
            <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 4px;">
                <span class="badge">КАРТА ПАЦИЕНТА</span>
                <div style="display: flex; gap: 8px;">
                    {{if .IsAdmin}}
                        <button onclick="try { window.Telegram.WebApp.openTelegramLink('https://t.me/{{.BotUsername}}?start=manual_{{.TelegramID}}'); } catch(e) { window.open('https://t.me/{{.BotUsername}}?start=manual_{{.TelegramID}}', '_blank'); }" class="btn-primary btn-admin">➕ Записать</button>
                    {{else}}
                        <a href="https://t.me/{{.BotUsername}}?start=book" class="btn-primary">🗓 Записаться</a>
                    {{end}}
                </div>
            </div>
            <h1>{{.Name}}</h1>
            <div class="stat-grid">
                <div class="stat-card">
                    <div class="stat-desc">Прогресс</div>
                    <div class="stat-val stat-val-large">{{.TotalVisits}} <span style="font-size: 12px; font-weight: 400; color: var(--text-muted); padding-left: 2px;">виз.</span></div>
                    <div class="stat-sub">Первый: {{.FirstVisit}}</div>
                </div>
                <div class="stat-card">
                    <div class="stat-desc">Текущая услуга</div>
                    <div class="stat-val" style="font-size: 13px;">{{.CurrentService}}</div>
                    <div class="stat-sub">{{.GeneratedAt}}</div>
                </div>
                {{if .FutureAppointments}}
                <div class="stat-card next-appt">
                    <div class="stat-desc">Следующий прием</div>
                    <div class="stat-val">{{(index .FutureAppointments 0).Date}}</div>
                    <div class="stat-sub">Ждем вас ❤️</div>
                </div>
                {{else}}
                <div class="stat-card">
                    <div class="stat-desc">Последний визит</div>
                    <div class="stat-val">{{.LastVisit}}</div>
                    <div class="stat-sub">Запишитесь снова</div>
                </div>
                {{end}}
            </div>
        </div>
    </header>
    <main class="main-container">
        <!-- UPCOMING APPOINTMENTS -->
        {{if .FutureAppointments}}
        <section class="upcoming">
            <h2 class="collapsible-header" onclick="toggleSection(this)" onkeydown="handleKey(event, this)" role="button" tabindex="0" aria-expanded="true">Будущие записи</h2>
            <div class="appt-list collapsible-content">
                {{range .FutureAppointments}}
                <div class="appt-item">
                    <div>
                        <div style="font-weight: 700; font-size: 15px;">{{.Date}}</div>
                        <div style="font-size: 12px; color: var(--text-muted);">{{.Service}}</div>
                    </div>
                    <div>
                        {{if .CanCancel}}
                            <button class="btn-cancel" type="button" onclick="cancelAppointment(event, '{{.ID}}', this)">Отменить</button>
                        {{else}}
                            <a href="https://t.me/VeraFethiye" class="contact-vera" aria-label="Написать Вере в Telegram">💬 Написать Вере</a>
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
            <h2 class="collapsible-header" onclick="toggleSection(this)" onkeydown="handleKey(event, this)" role="button" tabindex="0" aria-expanded="true">История болезни</h2>
            <div class="notes-content collapsible-content">
                {{if .TherapistNotes}}
                    {{.TherapistNotes}}
                {{else}}
                    <div class="empty-state">
                        <span class="empty-state-icon">📝</span>
                        <div class="empty-state-text">Записей пока нет</div>
                        <div class="empty-state-sub">Вера добавит важную информацию после вашего следующего визита.</div>
                    </div>
                {{end}}
            </div>
        </section>

        <section class="history collapsed">
            <h2 class="collapsible-header" onclick="toggleSection(this)" onkeydown="handleKey(event, this)" role="button" tabindex="0" aria-expanded="false" aria-controls="history-content">История посещений</h2>
            <div id="history-content" class="collapsible-content appt-list">
                {{if .RecentVisits}}
                    {{range .RecentVisits}}
                    <div class="appt-item">
                        <div style="font-weight: 600; font-size: 14px;">{{.Date}}</div>
                        <div style="font-size: 13px; color: var(--text-muted);">{{.Service}}</div>
                    </div>
                    {{end}}
                {{else}}
                    <div class="empty-state" style="border: none; padding: 20px 0;">
                        <span class="empty-state-icon">🕒</span>
                        <div class="empty-state-text">История пуста</div>
                    </div>
                {{end}}
            </div>
        </section>

        {{if .VoiceTranscripts}}
        <section>
            <h2 class="collapsible-header" onclick="toggleSection(this)" onkeydown="handleKey(event, this)" role="button" tabindex="0" aria-expanded="true">Расшифровки Консультаций</h2>
            <div class="notes-content collapsible-content" style="font-style: italic; color: #64748b;">{{.VoiceTranscripts}}</div>
        </section>
        {{end}}

        <section class="collapsed">
            <h2 class="collapsible-header" onclick="toggleSection(this)" onkeydown="handleKey(event, this)" role="button" tabindex="0" aria-expanded="false" aria-controls="docs-content">Документы и Снимки</h2>
            <div id="docs-content" class="collapsible-content doc-list">
                {{if .DocGroups}}
                    {{range .DocGroups}}
                    <div class="doc-group">
                        <div class="doc-item" onclick="this.parentElement.querySelector('.doc-files').classList.toggle('open')">
                            <div class="doc-info">
                                <div style="font-weight: 600;">{{.Name}}</div>
                                <div class="doc-stat">Количество: {{.Count}}</div>
                            </div>
                            <div class="doc-latest">
                                <div>Последний:</div>
                                <div>{{.Latest}}</div>
                            </div>
                        </div>
                        <div class="doc-files">
                            {{range .Files}}
                            <div class="media-item" onclick="openMedia('/api/media/{{.ID}}', '{{.FileType}}')">
                                {{if or (eq .FileType "photo") (eq .FileType "image") (eq .FileType "scan")}}
                                    <img src="/api/media/{{.ID}}" loading="lazy" alt="Image">
                                {{else if eq .FileType "video"}}
                                    <div class="file-icon-box">
                                        <div class="file-icon">📹</div>
                                        <div class="file-date">{{.CreatedAt.Format "02.01"}}</div>
                                    </div>
                                {{else if eq .FileType "voice"}}
                                    <div class="file-icon-box">
                                        <div class="file-icon">🎤</div>
                                        <div class="file-date">{{.CreatedAt.Format "02.01"}}</div>
                                    </div>
                                {{else}}
                                    <div class="file-icon-box">
                                        <div class="file-icon">📄</div>
                                        <div class="file-date">{{.CreatedAt.Format "02.01"}}</div>
                                    </div>
                                {{end}}
                            </div>
                            {{end}}
                        </div>
                    </div>
                    {{end}}
                {{else}}
                    <div class="empty-state" style="border: none; padding: 20px 0;">
                        <span class="empty-state-icon">📂</span>
                        <div class="empty-state-text">Файлов пока нет</div>
                    </div>
                {{end}}
            </div>
        </section>
        <footer class="footer">Vera Massage Bot {{.BotVersion}}<br>Professional Medical Record Hub</footer>
        
        <!-- Lightbox Overlay -->
        <div id="lightbox" class="lightbox" onclick="if(event.target === this) closeLightbox()">
            <button class="lightbox-close" onclick="closeLightbox()">✕</button>
            <div id="lightbox-media"></div>
        </div>
    </main>
    <script>
        function openMedia(url, type) {
            const lb = document.getElementById('lightbox');
            const container = document.getElementById('lightbox-media');
            container.innerHTML = '';
            
            // Normalize type
            if (type === 'image' || type === 'photo' || type === 'scan') type = 'image';
            
            try {
                if (type === 'video') {
                    const vid = document.createElement('video');
                    vid.src = url;
                    vid.controls = true;
                    vid.autoplay = true;
                    vid.playsInline = true; // Important for iOS
                    vid.className = 'lightbox-content';
                    container.appendChild(vid);
                } else if (type === 'voice' || type === 'audio') {
                    const audio = document.createElement('audio');
                    audio.src = url;
                    audio.controls = true;
                    audio.autoplay = true;
                    audio.className = 'lightbox-content';
                    // Add a visual placeholder for audio
                    const icon = document.createElement('div');
                    icon.innerHTML = '🎤 Аудиозапись';
                    icon.style.color = 'white';
                    icon.style.fontSize = '24px';
                    icon.style.marginBottom = '20px';
                    container.insertBefore(icon, audio);
                } else if (type === 'image') {
                    const img = document.createElement('img');
                    img.src = url;
                    img.className = 'lightbox-content';
                    container.appendChild(img);
                } else {
                    // Fallback for documents: try to open in new window (might fail auth, but best effort)
                    // Or ideally showing a "Cannot preview" message
                     window.open(url, '_blank');
                     return;
                }
                
                lb.classList.add('visible');
            } catch (e) {
                console.error("Error opening media:", e);
                window.Telegram.WebApp.showAlert("Ошибка при открытии файла");
            }
        }

        function closeLightbox() {
            const lb = document.getElementById('lightbox');
            lb.classList.remove('visible');
            const video = lb.querySelector('video');
            if (video) video.pause();
            const audio = lb.querySelector('audio');
            if (audio) audio.pause();
            
            setTimeout(() => {
                document.getElementById('lightbox-media').innerHTML = '';
            }, 300);
        }
    </script>
</body>
</html>
`

const adminSearchTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Админ Панель: Поиск Пациента</title>
    <script src="https://telegram.org/js/telegram-web-app.js"></script>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; padding: 20px; background-color: var(--tg-theme-bg-color, #fff); color: var(--tg-theme-text-color, #000); }
        .container { max-width: 600px; margin: 0 auto; }
        h1 { font-size: 20px; margin-bottom: 20px; }
        .search-box { display: flex; gap: 10px; margin-bottom: 20px; }
        input { flex: 1; padding: 12px; border-radius: 8px; border: 1px solid #ccc; font-size: 16px; }
        button { padding: 12px 20px; background-color: var(--tg-theme-button-color, #3390ec); color: var(--tg-theme-button-text-color, #fff); border: none; border-radius: 8px; cursor: pointer; font-size: 16px; font-weight: bold; }
        .results { display: flex; flex-direction: column; gap: 10px; }
        .patient-card { padding: 15px; background: var(--tg-theme-secondary-bg-color, #f5f5f5); border-radius: 12px; cursor: pointer; transition: background 0.2s; border: 1px solid transparent; }
        .patient-card:hover { border-color: var(--tg-theme-button-color, #3390ec); }
        .patient-name { font-weight: bold; font-size: 16px; color: var(--tg-theme-text-color, #000); }
        .patient-info { font-size: 13px; color: var(--tg-theme-hint-color, #888); margin-top: 4px; }
        .btn-row { margin-top: 12px; display: flex; gap: 8px; }
        .btn-action { border: none; border-radius: 8px; padding: 8px 12px; font-size: 12px; font-weight: 700; cursor: pointer; color: white; flex: 1; text-align: center; text-decoration: none; }
        .btn-create { background-color: #10b981; }
        .btn-view { background-color: var(--tg-theme-button-color, #3390ec); }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔍 Поиск Пациента</h1>
        <div class="search-box">
            <input type="text" id="query" placeholder="Имя или ID..." onkeypress="handleEnter(event)">
            <button onclick="search()">Найти</button>
        </div>
        <div id="results" class="results"></div>
    </div>

    <script>
        const tg = window.Telegram.WebApp;
        tg.expand();

        function handleEnter(e) {
            if (e.key === 'Enter') search();
        }

        async function search() {
            const query = document.getElementById('query').value;
            // if (!query) return; // Allow empty query for list all
            
            const btn = document.querySelector('button');
            const originalText = btn.innerText;
            btn.innerText = '⌛';
            
            try {
                const resp = await fetch('/api/search?q=' + encodeURIComponent(query), {
                    headers: { 'X-Telegram-Init-Data': tg.initData }
                });
                const data = await resp.json();
                
                const container = document.getElementById('results');
                container.innerHTML = '';

                if (!data || data.length === 0) {
                    container.innerHTML = '<div style="text-align:center;color:#888">Ничего не найдено</div>';
                    return;
                }

                data.forEach(p => {
                    const el = document.createElement('div');
                    el.className = 'patient-card';
                    el.onclick = () => viewPatient(p.telegram_id);
                    el.innerHTML = '<div class="patient-name">' + p.name + '</div>' +
                        '<div class="patient-info">ID: ' + p.telegram_id + ' • Визитов: ' + p.total_visits + '</div>' +
                        '<div class="btn-row">' +
                            '<button onclick="event.stopPropagation(); try { window.Telegram.WebApp.openTelegramLink(\'https://t.me/{{.BotUsername}}?start=manual_\' + \'' + p.telegram_id + '\'); } catch(e) { window.open(\'https://t.me/{{.BotUsername}}?start=manual_\' + \'' + p.telegram_id + '\', \'_blank\'); }" class="btn-action btn-create">➕ Новая запись</button>' +
                            '<button onclick="event.stopPropagation(); viewPatient(\'' + p.telegram_id + '\')" class="btn-action btn-view">📄 Карта</button>' +
                        '</div>';
                    container.appendChild(el);
                });
            } catch (e) {
                alert('Ошибка поиска: ' + e.message);
            } finally {
                btn.innerText = originalText;
            }
        }
        
        // Auto-load list on start
        window.addEventListener('DOMContentLoaded', search);

        function viewPatient(id) {
            // Reload page with ID param to view their card
            const url = new URL(window.location.href);
            url.searchParams.set('id', id);
            // We keep initData/token if present
            window.location.href = url.toString();
        }
    </script>
</body>
</html>
`
