package storage

const medicalRecordTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <title>Медицинская карта - {{.Name}}</title>
    <script src="https://telegram.org/js/telegram-web-app.js"></script>
    <script>
        try {
            window.addEventListener('DOMContentLoaded', (event) => {
                const tg = window.Telegram.WebApp;
                if (tg && tg.expand) {
                    tg.expand();
                    if (tg.themeParams && tg.themeParams.secondary_bg_color) {
                        // Theming logic for TWA integration
                        document.body.style.backgroundColor = tg.themeParams.secondary_bg_color;
                    }
                }
            });
        } catch (e) { console.error('TWA init failed:', e); }

        function printToPDF() {
            window.print();
        }
    </script>
    <style>
        :root {
            --primary: #2563eb;
            --primary-dark: #1d4ed8;
            --bg-page: #f8fafc;
            --bg-card: #ffffff;
            --text-main: #0f172a;
            --text-muted: #64748b;
            --border: #e2e8f0;
            --accent-soft: #eff6ff;
            --accent-indigo: #4f46e5;
        }

        * { box-sizing: border-box; -webkit-print-color-adjust: exact; }

        body {
            background-color: var(--bg-page);
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            margin: 0;
            padding: 12px;
            color: var(--text-main);
            line-height: 1.5;
            -webkit-font-smoothing: antialiased;
        }

        .container {
            max-width: 850px;
            margin: 0 auto;
            background: var(--bg-card);
            border-radius: 16px;
            box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.05), 0 8px 10px -6px rgba(0, 0, 0, 0.05);
            overflow: hidden;
            border: 1px solid var(--border);
        }

        /* ACTIONS BAR - Non-printable */
        .actions-bar {
            display: flex;
            justify-content: flex-end;
            padding: 12px 20px;
            background: #ffffff;
            border-bottom: 1px solid var(--border);
        }
        .btn-print {
            background: var(--primary);
            color: white;
            border: none;
            padding: 8px 16px;
            border-radius: 8px;
            font-weight: 600;
            font-size: 13px;
            cursor: pointer;
            display: flex;
            align-items: center;
            gap: 6px;
            transition: background 0.2s;
        }
        .btn-print:hover { background: var(--primary-dark); }

        /* HEADER */
        .header {
            padding: 30px;
            background: linear-gradient(to right, #ffffff, #f1f5f9);
            border-bottom: 3px solid var(--primary);
            display: flex;
            flex-wrap: wrap;
            gap: 24px;
            align-items: flex-start;
        }

        .patient-identity { flex: 1; min-width: 280px; }
        .patient-name { 
            font-size: 32px; 
            font-weight: 800; 
            letter-spacing: -0.025em; 
            margin: 0 0 4px 0;
            color: #000;
        }
        .patient-meta {
            font-size: 13px;
            color: var(--text-muted);
            font-weight: 500;
            display: flex;
            gap: 12px;
            align-items: center;
        }
        .id-badge {
            background: var(--accent-soft);
            color: var(--primary);
            padding: 2px 8px;
            border-radius: 4px;
            font-weight: 700;
            font-family: monospace;
        }

        .header-stats {
            display: flex;
            gap: 12px;
        }
        .stat-item {
            background: white;
            border: 1px solid var(--border);
            padding: 12px 20px;
            border-radius: 12px;
            text-align: center;
            min-width: 110px;
        }
        .stat-label { font-size: 10px; text-transform: uppercase; font-weight: 700; color: var(--text-muted); margin-bottom: 2px; }
        .stat-value { font-size: 24px; font-weight: 800; color: var(--primary); line-height: 1; }

        /* MAIN LAYOUT */
        .content {
            padding: 30px;
            display: grid;
            grid-template-columns: 1fr;
            gap: 30px;
        }

        @media (min-width: 768px) {
            .content { grid-template-columns: 1fr 280px; }
        }

        .main-section { display: flex; flex-direction: column; gap: 32px; }
        
        .section-header {
            display: flex;
            align-items: center;
            gap: 10px;
            margin-bottom: 16px;
            padding-bottom: 8px;
            border-bottom: 1px solid var(--border);
        }
        .section-title {
            font-size: 14px;
            font-weight: 800;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            color: var(--text-muted);
        }

        /* CARDS / NOTES */
        .note-card {
            background: var(--accent-soft);
            border-left: 4px solid var(--primary);
            padding: 24px;
            border-radius: 8px;
        }
        .note-content {
            white-space: pre-wrap;
            font-size: 16px;
            color: #1e293b;
            margin: 0;
            line-height: 1.6;
        }

        .transcript-box {
            background: #fdfdfd;
            border: 1px solid var(--border);
            padding: 20px;
            border-radius: 12px;
            font-style: italic;
            color: #475569;
            font-size: 14px;
        }

        /* HISTORY TABLE */
        .history-table {
            width: 100%;
            border-collapse: collapse;
            font-size: 14px;
        }
        .history-row {
            border-bottom: 1px solid #f1f5f9;
        }
        .history-row:last-child { border-bottom: none; }
        .history-cell { padding: 12px 0; }
        .h-date { color: var(--text-muted); width: 140px; font-weight: 500; }
        .h-service { font-weight: 600; color: var(--text-main); flex: 1; }
        .h-action { text-align: right; }

        .cal-badge {
            display: inline-flex;
            align-items: center;
            gap: 4px;
            padding: 4px 8px;
            background: var(--accent-soft);
            color: var(--primary);
            text-decoration: none;
            border-radius: 6px;
            font-size: 11px;
            font-weight: 700;
            transition: all 0.2s;
            border: 1px solid transparent;
        }
        .cal-badge:hover {
            background: var(--primary);
            color: white;
            border-color: var(--primary);
        }

        /* SIDEBAR COMPONENTS */
        .sidebar { display: flex; flex-direction: column; gap: 24px; }
        
        .program-card {
            background: #1e293b;
            color: white;
            padding: 20px;
            border-radius: 12px;
            position: relative;
            overflow: hidden;
        }
        .program-label { font-size: 10px; text-transform: uppercase; opacity: 0.7; font-weight: 700; margin-bottom: 4px; }
        .program-name { font-size: 18px; font-weight: 700; position: relative; z-index: 1; }
        .program-card::after {
            content: "";
            position: absolute;
            top: -20px; right: -20px;
            width: 80px; height: 80px;
            background: rgba(255,255,255,0.05);
            border-radius: 50%;
        }

        .doc-list { list-style: none; padding: 0; margin: 0; }
        .doc-item {
            display: flex;
            align-items: center;
            gap: 10px;
            padding: 10px;
            background: #f8fafc;
            border: 1px solid var(--border);
            border-radius: 8px;
            margin-bottom: 8px;
            font-size: 13px;
            font-weight: 500;
        }
        .doc-icon {
            width: 8px; height: 8px; border-radius: 50%;
            background: var(--primary);
        }

        /* FOOTER */
        .footer {
            padding: 24px;
            text-align: center;
            border-top: 1px solid var(--border);
            background: #fafafa;
        }
        .copyright { font-size: 11px; font-weight: 700; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.1em; }

        @media print {
            body { background: white; padding: 0; }
            .container { border: none; box-shadow: none; width: 100%; max-width: 100%; }
            .actions-bar { display: none; }
            .note-card { background: white; border: 1px solid #eee; border-left: 4px solid #000; }
            .stat-item { border: 2px solid #000; }
            .btn-print, .cal-badge { display: none; }
        }
    </style>
</head>
<body>
    <div class="container">
        <!-- Action bar is only visible in browser/TWA -->
        <div class="actions-bar">
            <button class="btn-print" onclick="printToPDF()">
                <span>📄 Сохранить PDF</span>
            </button>
        </div>

        <header class="header">
            <div class="patient-identity">
                <h1 class="patient-name">{{.Name}}</h1>
                <div class="patient-meta">
                    <span>ID: <span class="id-badge">{{.TelegramID}}</span></span>
                    <span>•</span>
                    <span>Vera Bot {{.BotVersion}}</span>
                </div>
            </div>
            <div class="header-stats">
                <div class="stat-item">
                    <div class="stat-label">Визиты</div>
                    <div class="stat-value">{{.TotalVisits}}</div>
                </div>
            </div>
        </header>

        <div class="content">
            <div class="main-section">
                <!-- CLINICAL NOTES -->
                <section>
                    <div class="section-header">
                        <span class="section-title">Анамнез и заметки терапевта</span>
                    </div>
                    <div class="note-card">
                        <p class="note-content">{{.TherapistNotes}}</p>
                    </div>
                </section>

                <!-- TRANSCRIPTS IF ANY -->
                {{if .VoiceTranscripts}}
                <section>
                    <div class="section-header">
                        <span class="section-title">Расшифровки консультаций</span>
                    </div>
                    <div class="transcript-box">
                        <p class="note-content">{{.VoiceTranscripts}}</p>
                    </div>
                </section>
                {{end}}

                <!-- VISIT HISTORY -->
                <section>
                    <div class="section-header">
                        <span class="section-title">История посещений</span>
                    </div>
                    <table class="history-table">
                        <tr class="history-row">
                            <td class="history-cell h-date">{{.LastVisit}}</td>
                            <td class="history-cell h-service">{{.CurrentService}} (Последний визит)</td>
                            <td class="history-cell h-action">
                                {{if .ShowLastVisitLink}}
                                <a href="{{.LastVisitLink}}" target="_blank" class="cal-badge">
                                    <span>📅</span> Календарь
                                </a>
                                {{end}}
                            </td>
                        </tr>
                        <tr class="history-row">
                            <td class="history-cell h-date">{{.FirstVisit}}</td>
                            <td class="history-cell h-service">{{.CurrentService}} (Первый визит)</td>
                            <td class="history-cell h-action">
                                {{if .ShowFirstVisitLink}}
                                <a href="{{.FirstVisitLink}}" target="_blank" class="cal-badge">
                                    <span>📅</span> Календарь
                                </a>
                                {{end}}
                            </td>
                        </tr>
                    </table>
                </section>
            </div>

            <!-- SIDEBAR -->
            <aside class="sidebar">
                <section>
                    <div class="section-header">
                        <span class="section-title">Программа</span>
                    </div>
                    <div class="program-card">
                        <div class="program-label">Текущий курс</div>
                        <div class="program-name">{{.CurrentService}}</div>
                    </div>
                </section>

                <section>
                    <div class="section-header">
                        <span class="section-title">Документация</span>
                    </div>
                    <ul class="doc-list">
                        {{range .Documents}}
                        <li class="doc-item">
                            <div class="doc-icon" style="background: {{if .IsVoice}}#a855f7{{else}}#2563eb{{end}}"></div>
                            {{.Name}}
                        </li>
                        {{else}}
                        <p style="font-size: 13px; color: #94a3b8; font-style: italic;">Список пуст</p>
                        {{end}}
                    </ul>
                </section>

                <section style="margin-top: auto; padding-top: 20px;">
                    <div style="font-size: 11px; color: var(--text-muted); border-top: 1px solid var(--border); padding-top: 10px;">
                        <strong>Сгенерировано:</strong><br>
                        {{.GeneratedAt}}
                    </div>
                </section>
            </aside>
        </div>

        <footer class="footer">
            <div class="copyright">Vera Massage Bot • Professional Medical Records</div>
        </footer>
    </div>
</body>
</html>
`
