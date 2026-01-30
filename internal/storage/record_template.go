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
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=Outfit:wght@500;700&display=swap" rel="stylesheet">
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

            if (confirm("Вы уверены, что хотите отменить эту запись?")) {
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
                        tg.showAlert("Ошибка: " + (result.error || "Не удалось отменить запись."));
                    }
                } catch (e) {
                    tg.showAlert("Ошибка сети при отмене записи.");
                }
            }
        }

        function updateCountdown() {
            const nextUnix = {{.NextApptUnix}};
            if (nextUnix === 0) return;

            const now = Math.floor(Date.now() / 1000);
            const diff = nextUnix - now;
            const el = document.getElementById('countdown-timer');
            const container = document.getElementById('countdown-container');
            
            if (!el || !container) return;

            if (diff <= 0) {
                container.style.display = 'none';
                return;
            }

            const days = Math.floor(diff / 86400);
            const hours = Math.floor((diff % 86400) / 3600);
            const mins = Math.floor((diff % 3600) / 60);

            let str = "";
            if (days > 0) str += days + "д ";
            if (hours > 0 || days > 0) str += hours + "ч ";
            str += mins + "м";
            
            el.innerText = str;
            container.style.display = 'flex';
        }

        window.addEventListener('DOMContentLoaded', () => {
            const tg = window.Telegram.WebApp;
            if (tg) { 
                tg.expand(); 
                tg.ready(); 
                tg.setHeaderColor('#ffffff'); // Match header bg
            }
            
            updateCountdown();
            setInterval(updateCountdown, 60000);
            
            // Stagger animation for cards
            const cards = document.querySelectorAll('section');
            cards.forEach((card, index) => {
                setTimeout(() => {
                    card.style.opacity = '1';
                    card.style.transform = 'translateY(0)';
                }, 100 * index);
            });
        });
    </script>
    <style>
        :root {
            --primary: #0284c7; /* Sky 600 - Medical Blue */
            --primary-light: #e0f2fe; /* Sky 100 */
            --bg-page: #f1f5f9; /* Slate 100 */
            --bg-card: #ffffff;
            --text-main: #0f172a; /* Slate 900 */
            --text-secondary: #64748b; /* Slate 500 */
            --border: #e2e8f0; /* Slate 200 */
            --danger: #ef4444; 
            --danger-bg: #fee2e2;
            --radius-lg: 16px;
            --radius-md: 12px;
            --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
            --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1);
        }

        * { box-sizing: border-box; -webkit-tap-highlight-color: transparent; }
        body { 
            background-color: var(--bg-page); 
            font-family: 'Inter', system-ui, -apple-system, sans-serif; 
            margin: 0; 
            padding: 0; 
            color: var(--text-main); 
            line-height: 1.5;
            padding-bottom: 40px;
        }

        /* Modern Glass Header */
        header {
            background: rgba(255, 255, 255, 0.9);
            backdrop-filter: blur(10px);
            -webkit-backdrop-filter: blur(10px);
            position: sticky;
            top: 0;
            z-index: 100;
            border-bottom: 1px solid var(--border);
            padding: 20px 24px 24px;
            box-shadow: var(--shadow-sm);
        }

        .header-top {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 16px;
        }

        h1 {
            font-family: 'Outfit', sans-serif;
            font-size: 26px;
            font-weight: 700;
            margin: 0;
            color: var(--text-main);
            letter-spacing: -0.02em;
            line-height: 1.2;
        }

        .patient-id {
            font-size: 11px;
            color: var(--text-secondary);
            font-weight: 500;
            margin-top: 4px;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        .badge {
            background: var(--text-main);
            color: white;
            font-size: 10px;
            font-weight: 700;
            padding: 4px 8px;
            border-radius: 6px;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        /* Stats Grid */
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 12px;
        }

        .stat-item {
            background: var(--bg-page);
            padding: 12px;
            border-radius: var(--radius-md);
            text-align: center;
        }

        .stat-value {
            font-family: 'Outfit', sans-serif;
            font-size: 18px;
            font-weight: 700;
            color: var(--primary);
            line-height: 1.2;
        }

        .stat-label {
            font-size: 10px;
            color: var(--text-secondary);
            margin-top: 4px;
            font-weight: 500;
        }

        /* Main Content */
        main {
            padding: 20px;
            max-width: 600px;
            margin: 0 auto;
        }

        section {
            background: var(--bg-card);
            border-radius: var(--radius-lg);
            padding: 20px;
            margin-bottom: 20px;
            border: 1px solid var(--border);
            box-shadow: var(--shadow-sm);
            
            /* Animation Initial State */
            opacity: 0;
            transform: translateY(10px);
            transition: opacity 0.4s ease, transform 0.4s ease;
        }

        h2 {
            font-family: 'Outfit', sans-serif;
            font-size: 15px;
            font-weight: 700;
            color: var(--text-main);
            margin: 0 0 16px 0;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        /* List Items */
        .list-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 12px 0;
            border-bottom: 1px solid var(--border);
        }

        .list-item:last-child {
            border-bottom: none;
            padding-bottom: 0;
        }
        
        .list-item:first-child {
            padding-top: 0;
        }

        .item-main {
            font-weight: 600;
            font-size: 14px;
            color: var(--text-main);
        }

        .item-sub {
            font-size: 12px;
            color: var(--text-secondary);
            margin-top: 2px;
        }

        /* Buttons & Actions */
        .btn-cancel {
            background: var(--danger-bg);
            color: var(--danger);
            border: none;
            padding: 6px 12px;
            border-radius: 8px;
            font-size: 12px;
            font-weight: 600;
            cursor: pointer;
            transition: opacity 0.2s;
        }

        .btn-cancel:active {
            opacity: 0.7;
        }

        .btn-action {
            display: inline-block;
            background: var(--primary-light);
            color: var(--primary);
            text-decoration: none;
            padding: 6px 12px;
            border-radius: 8px;
            font-size: 12px;
            font-weight: 600;
        }

        /* Countdown Banner */
        #countdown-container {
            background: var(--primary);
            color: white;
            border-radius: var(--radius-md);
            padding: 12px 16px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
            box-shadow: var(--shadow-md);
        }

        .countdown-label {
            font-size: 12px;
            font-weight: 500;
            opacity: 0.9;
        }

        .countdown-time {
            font-family: 'Outfit', sans-serif;
            font-size: 16px;
            font-weight: 700;
        }

        /* Notes & Typography */
        .notes-text {
            font-size: 14px;
            line-height: 1.6;
            color: #334155;
            white-space: pre-wrap;
        }

        .notes-text h1, .notes-text h2 {
            font-size: 16px;
            margin-top: 16px;
            margin-bottom: 8px;
            color: var(--text-main);
        }

        .empty-state {
            text-align: center;
            color: var(--text-secondary);
            font-size: 13px;
            padding: 16px 0;
        }

        /* Document List */
        .doc-row {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 12px;
            background: var(--bg-page);
            border-radius: var(--radius-md);
            margin-bottom: 8px;
            text-decoration: none;
            color: inherit;
        }

        .doc-name {
            font-weight: 600;
            font-size: 14px;
        }

        .doc-meta {
            font-size: 11px;
            color: var(--text-secondary);
            text-align: right;
        }

        footer {
            text-align: center;
            font-size: 11px;
            color: var(--text-secondary);
            margin-top: 32px;
            opacity: 0.6;
        }
    </style>
</head>
<body>
    <header>
        <div class="header-top">
            <div>
                <h1>{{.Name}}</h1>
                <div class="patient-id">ID: {{.TelegramID}}</div>
            </div>
            <div class="badge">V{{.BotVersion}}</div>
        </div>
        
        <div class="stats-grid">
            <div class="stat-item">
                <div class="stat-value">{{.TotalVisits}}</div>
                <div class="stat-label">Посещений</div>
            </div>
            <div class="stat-item">
                <div class="stat-value">{{.CurrentService}}</div>
                <div class="stat-label">Текущий курс</div>
            </div>
            <div class="stat-item">
                <div class="stat-value">{{.LastVisit}}</div>
                <div class="stat-label">Последний визит</div>
            </div>
        </div>
    </header>

    <main>
        <!-- Next Appointment Countdown -->
        {{if .NextApptUnix}}
        <div id="countdown-container" style="display: none;">
            <span class="countdown-label">До следующего приема</span>
            <span class="countdown-time" id="countdown-timer">Загрузка...</span>
        </div>
        {{end}}

        <!-- Future Appointments -->
        {{if .FutureAppointments}}
        <section>
            <h2>Предстоящие записи</h2>
            <div>
                {{range .FutureAppointments}}
                <div class="list-item">
                    <div>
                        <div class="item-main">{{.Date}}</div>
                        <div class="item-sub">{{.Service}}</div>
                    </div>
                    <div>
                        {{if .CanCancel}}
                            <button class="btn-cancel" onclick="cancelAppointment('{{.ID}}')">Отменить</button>
                        {{else}}
                            <a href="https://t.me/VeraFethiye" class="btn-action">Связаться</a>
                        {{end}}
                    </div>
                </div>
                {{end}}
            </div>
            <div style="font-size: 11px; color: var(--text-secondary); margin-top: 12px; text-align: center;">
                Бесплатная отмена возможна за 72 часа.
            </div>
        </section>
        {{end}}

        <!-- Medical History -->
        <section>
            <h2>История болезни</h2>
            <div class="notes-text">{{.TherapistNotes}}</div>
        </section>

        <!-- Past Visits -->
        {{if .RecentVisits}}
        <section>
            <h2>История посещений</h2>
            <div>
                {{range .RecentVisits}}
                <div class="list-item">
                    <div>
                        <div class="item-main">{{.Date}}</div>
                        <div class="item-sub">{{.Service}}</div>
                    </div>
                </div>
                {{end}}
            </div>
        </section>
        {{end}}

        <!-- Transcripts -->
        {{if .VoiceTranscripts}}
        <section>
            <h2>Расшифровки</h2>
            <div class="notes-text" style="font-style: italic; color: #475569;">{{.VoiceTranscripts}}</div>
        </section>
        {{end}}

        <!-- Documents -->
        <section>
            <h2>Документы</h2>
            <div style="display: flex; flex-direction: column; gap: 8px;">
                {{range .DocGroups}}
                <div class="doc-row">
                    <div>
                        <div class="doc-name">{{.Name}}</div>
                        <div class="item-sub">{{.Count}} файлов</div>
                    </div>
                    <div class="doc-meta">
                        <div>Обновлено</div>
                        <div>{{.Latest}}</div>
                    </div>
                </div>
                {{else}}
                <div class="empty-state">Документов пока нет</div>
                {{end}}
            </div>
        </section>

        <footer>
            Massage Bot Medical System<br>
            Secure Patient Portal
        </footer>
    </main>
</body>
</html>
`
