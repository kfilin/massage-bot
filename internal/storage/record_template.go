package storage

const medicalRecordTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <title>Медицинская карта - {{.Name}}</title>
    <script src="https://telegram.org/js/telegram-web-app.js"></script>
    <script>
        try {
            window.addEventListener('DOMContentLoaded', (event) => {
                const tg = window.Telegram.WebApp;
                if (tg && tg.expand) {
                    tg.expand();
                    if (tg.themeParams && tg.themeParams.secondary_bg_color) {
                        document.body.style.backgroundColor = tg.themeParams.secondary_bg_color;
                        document.body.style.color = tg.themeParams.text_color;
                    }
                }
            });
        } catch (e) { console.error('TWA script error:', e); }
    </script>
    <style>
        * { box-sizing: border-box; -webkit-print-color-adjust: exact; }
        
        body { 
            background-color: #f1f5f9; 
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            margin: 0;
            padding: 15px;
            line-height: 1.6;
            color: #1e293b;
        }

        @media (min-width: 768px) {
            body { padding: 40px; }
        }

        .page {
            max-width: 900px;
            margin: 0 auto;
            background: white;
            padding: 25px;
            border-radius: 12px;
            box-shadow: 0 4px 20px rgba(0,0,0,0.08);
        }

        @media (min-width: 768px) {
            .page { padding: 50px; }
        }

        /* Header Layout */
        .header {
            border-bottom: 4px solid #2563eb;
            padding-bottom: 25px;
            margin-bottom: 40px;
            display: flex;
            flex-wrap: wrap;
            justify-content: space-between;
            align-items: flex-end;
            gap: 20px;
        }

        .header-info { flex: 1; min-width: 250px; }
        .patient-name { font-size: 36px; font-weight: 800; color: #0f172a; margin: 0; letter-spacing: -1px; }
        .medical-id { font-size: 14px; color: #2563eb; font-weight: 700; margin: 5px 0; text-transform: uppercase; }
        .system-tag { font-size: 11px; color: #64748b; }

        .header-stats { text-align: right; }
        .stat-badge { background: #2563eb; color: white; padding: 15px 30px; border-radius: 12px; display: inline-block; text-align: center; }
        .stat-label { font-size: 11px; text-transform: uppercase; font-weight: 800; margin: 0; opacity: 0.9; }
        .stat-value { font-size: 28px; font-weight: 900; margin: 0; line-height: 1; }
        .generated-date { font-size: 11px; color: #64748b; margin-top: 10px; font-weight: 600; text-transform: uppercase; }

        /* General Sections */
        .section { margin-bottom: 45px; }
        .section-title { font-size: 12px; font-weight: 800; color: #64748b; text-transform: uppercase; letter-spacing: 1.5px; margin-bottom: 15px; display: block; border-bottom: 1px solid #f1f5f9; padding-bottom: 5px; }

        /* Simple flex columns for main content */
        .content-layout { display: flex; flex-wrap: wrap; gap: 40px; }
        .sidebar { flex: 0 0 280px; }
        .main { flex: 1; min-width: 300px; }

        /* Components */
        .card { background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 10px; padding: 20px; }
        .service-name { font-size: 18px; font-weight: 700; color: #0f172a; margin: 0; }
        .service-status { font-size: 12px; color: #2563eb; font-weight: 700; text-transform: uppercase; margin-top: 4px; }

        .note-container { 
            background: #eff6ff; 
            border: 1px solid #bfdbfe; 
            border-left: 6px solid #2563eb; 
            padding: 25px; 
            border-radius: 4px 12px 12px 4px; 
            font-size: 15px;
            color: #1e293b;
        }
        .note-text { margin: 0; white-space: pre-wrap; }

        .transcript-container { 
            background: #f8fafc; 
            border: 1px solid #e2e8f0; 
            border-left: 6px solid #64748b; 
            padding: 20px; 
            border-radius: 4px 12px 12px 4px; 
            font-size: 14px;
            color: #475569;
            font-style: italic;
        }

        /* Lists */
        .doc-item { display: flex; align-items: center; gap: 12px; padding: 12px; background: white; border: 1px solid #e2e8f0; border-radius: 8px; margin-bottom: 10px; font-size: 13px; }
        .doc-dot { width: 10px; height: 10px; border-radius: 50%; }

        /* Table/Ledger */
        .ledger { border: 1px solid #e2e8f0; border-radius: 12px; overflow: hidden; }
        .ledger-row { display: flex; border-bottom: 1px solid #f1f5f9; padding: 15px 20px; }
        .ledger-row:last-child { border-bottom: none; }
        .ledger-header { background: #f8fafc; font-size: 11px; font-weight: 800; color: #64748b; text-transform: uppercase; }
        .l-date { flex: 0 0 160px; font-weight: 600; }
        .l-service { flex: 1; }

        /* Footer */
        .footer { margin-top: 60px; padding-top: 30px; border-top: 1px solid #f1f5f9; text-align: center; }
        .footer-text { font-size: 11px; color: #94a3b8; font-weight: 700; text-transform: uppercase; letter-spacing: 1px; }

        @media print {
            body { background: white; padding: 0; }
            .page { box-shadow: none; border: none; padding: 0; width: 100%; max-width: 100%; }
            .stat-badge { border: 2px solid #2563eb; color: #2563eb !important; background: transparent !important; }
            .stat-value, .stat-label { color: #2563eb !important; }
        }
    </style>
</head>
<body>
    <div class="page">
        <header class="header">
            <div class="header-info">
                <h1 class="patient-name">{{.Name}}</h1>
                <p class="medical-id">Медицинская карта ID {{.TelegramID}}</p>
                <p class="system-tag">Система: Vera Massage Bot v.2.1</p>
            </div>
            <div class="header-stats">
                <div class="stat-badge">
                    <p class="stat-label">Всего визитов</p>
                    <p class="stat-value">{{.TotalVisits}}</p>
                </div>
                <p class="generated-date">Дата: {{.GeneratedAt}}</p>
            </div>
        </header>

        <div class="content-layout">
            <div class="sidebar">
                <div class="section">
                    <span class="section-title">Текущий курс</span>
                    <div class="card">
                        <p class="service-name">{{.CurrentService}}</p>
                        <p class="service-status">Активная программа</p>
                    </div>
                </div>

                <div class="section">
                    <span class="section-title">Документация</span>
                    {{range .Documents}}
                    <div class="doc-item">
                        <div class="doc-dot" style="background: {{if .IsVoice}}#a855f7{{else}}#2563eb{{end}}"></div>
                        {{.Name}}
                    </div>
                    {{else}}
                    <p style="font-size: 13px; color: #94a3b8; font-style: italic;">Документы не загружены.</p>
                    {{end}}
                </div>
            </div>

            <div class="main">
                <div class="section">
                    <span class="section-title">Клинические заметки</span>
                    <div class="note-container">
                        <p class="note-text">{{.TherapistNotes}}</p>
                    </div>
                </div>

                {{if .VoiceTranscripts}}
                <div class="section">
                    <span class="section-title">Расшифровки голос. сообщений</span>
                    <div class="transcript-container">
                        <p class="note-text">{{.VoiceTranscripts}}</p>
                    </div>
                </div>
                {{end}}

                <div class="section">
                    <span class="section-title">История визитов</span>
                    <div class="ledger">
                        <div class="ledger-row ledger-header">
                            <div class="l-date">Дата и время</div>
                            <div class="l-service">Услуга</div>
                        </div>
                        <div class="ledger-row">
                            <div class="l-date">{{.LastVisit}}</div>
                            <div class="l-service">{{.CurrentService}} (Последнее посещение)</div>
                        </div>
                        <div class="ledger-row">
                            <div class="l-date">{{.FirstVisit}}</div>
                            <div class="l-service">{{.CurrentService}} (Первое посещение)</div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <footer class="footer">
            <p class="footer-text">Электронный медицинский документ • Vera Bot v.2.1</p>
        </footer>
    </div>
</body>
</html>
`
