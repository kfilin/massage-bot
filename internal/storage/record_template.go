package storage

const medicalRecordTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <title>Медицинская карта - {{.Name}}</title>
    <script src="https://telegram.org/js/telegram-web-app.js"></script>
    <script>
        window.addEventListener('DOMContentLoaded', (event) => {
            const tg = window.Telegram.WebApp;
            tg.expand();
            // Matching Telegram theme for Web view
            if (tg.themeParams.bg_color) {
                document.body.style.backgroundColor = tg.themeParams.bg_color;
                document.body.style.color = tg.themeParams.text_color;
            }
        });
    </script>
    <style>
        /* General Layout */
        * { box-sizing: border-box; }
        body { 
            background-color: #f1f5f9; 
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            margin: 0;
            padding: 2rem;
            line-height: 1.5;
            color: #1e293b;
            -webkit-print-color-adjust: exact;
        }
        .page {
            max-width: 850px;
            margin: 0 auto;
            background: white;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 10px 25px rgba(0,0,0,0.05);
        }
        
        /* Header */
        .header {
            border-bottom: 4px solid #2563eb;
            padding-bottom: 1.5rem;
            margin-bottom: 2rem;
            overflow: hidden; /* Clearfix */
        }
        .header-left { float: left; width: 70%; }
        .header-right { float: right; width: 30%; text-align: right; }
        
        .patient-name {
            font-size: 2.25rem;
            font-weight: 800;
            color: #0f172a;
            text-transform: uppercase;
            margin: 0;
        }
        .medical-id {
            font-size: 0.875rem;
            color: #2563eb;
            font-weight: 700;
            margin: 0.5rem 0 0 0;
            letter-spacing: 0.05em;
        }
        .system-tag { font-size: 11px; color: #94a3b8; margin: 0.25rem 0 0 0; }
        
        .badge {
            background: #2563eb;
            color: white;
            padding: 10px 20px;
            border-radius: 8px;
            display: inline-block;
        }
        .badge-label { font-size: 10px; text-transform: uppercase; font-weight: 700; margin: 0; opacity: 0.9; }
        .badge-value { font-size: 1.5rem; font-weight: 800; margin: 0; }
        .generated-at { font-size: 10px; color: #94a3b8; margin-top: 10px; font-weight: 600; text-transform: uppercase; }

        /* Grid */
        .dashboard-grid { width: 100%; margin-top: 2rem; clear: both; }
        .sidebar { width: 32%; float: left; }
        .main-content { width: 64%; float: right; }
        .dashboard-grid::after { content: ""; display: table; clear: both; }

        .section-label {
            font-size: 11px;
            font-weight: 700;
            color: #64748b;
            text-transform: uppercase;
            letter-spacing: 0.1em;
            margin-bottom: 0.75rem;
            display: block;
        }
        
        /* Cards */
        .info-card {
            background: #f8fafc;
            padding: 1.25rem;
            border-radius: 8px;
            border: 1px solid #e2e8f0;
            margin-bottom: 1.5rem;
        }
        .program-name { font-size: 1rem; font-weight: 700; color: #1e293b; margin: 0; }
        .program-sub { font-size: 11px; color: #3b82f6; font-weight: 700; text-transform: uppercase; margin-top: 4px; display: block; }

        .attachment-item {
            background: #f8fafc;
            border: 1px solid #e2e8f0;
            padding: 8px 12px;
            border-radius: 6px;
            margin-bottom: 8px;
            display: block;
            font-size: 11px;
            color: #475569;
        }
        .dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; margin-right: 8px; }
        .dot-blue { background: #3b82f6; }
        .dot-purple { background: #a855f7; }

        /* Clinical Status */
        .notes-box {
            background: #eff6ff;
            padding: 1.5rem;
            border-radius: 8px;
            border: 1px solid #bfdbfe;
            border-left: 6px solid #2563eb;
            margin-bottom: 2rem;
            page-break-inside: avoid;
        }
        .notes-text { font-size: 0.95rem; color: #1e293b; font-style: italic; margin: 0; }

        .transcript-box {
            background: #f8fafc;
            padding: 1.5rem;
            border-radius: 8px;
            border: 1px solid #e2e8f0;
            border-left: 6px solid #64748b;
            margin-bottom: 2rem;
            page-break-inside: avoid;
        }
        .transcript-text { font-size: 0.875rem; color: #475569; margin: 0; white-space: pre-wrap; }

        /* Table */
        .table-container { 
            border: 1px solid #e2e8f0; 
            border-radius: 8px; 
            overflow: hidden; 
            margin-bottom: 2rem;
        }
        table { width: 100%; border-collapse: collapse; text-align: left; font-size: 0.875rem; }
        th { background: #f8fafc; padding: 12px; font-weight: 700; font-size: 10px; color: #64748b; text-transform: uppercase; border-bottom: 1px solid #e2e8f0; }
        td { padding: 12px; border-bottom: 1px solid #f1f5f9; }

        /* Footer */
        .footer { margin-top: 3rem; text-align: center; border-top: 1px solid #f1f5f9; padding-top: 1.5rem; }
        .footer-text { font-size: 10px; color: #94a3b8; font-weight: 600; text-transform: uppercase; letter-spacing: 0.1em; }

        /* Print Overrides */
        @media print {
            body { background: white !important; padding: 0 !important; }
            .page { box-shadow: none !important; border: none !important; width: 100% !important; max-width: 100% !important; padding: 0 !important; }
            .badge { background: white !important; color: #2563eb !important; border: 2px solid #2563eb !important; box-shadow: none !important; }
            .badge-value, .badge-label { color: #2563eb !important; }
        }
    </style>
</head>
<body>
    <div class="page">
        <header class="header">
            <div class="header-left">
                <h1 class="patient-name">{{.Name}}</h1>
                <p class="medical-id">МЕДИЦИНСКАЯ КАРТА ID {{.TelegramID}}</p>
                <p class="system-tag">Система: Vera Massage Bot</p>
            </div>
            <div class="header-right">
                <div class="badge">
                    <p class="badge-label">Посещений</p>
                    <p class="badge-value">{{.TotalVisits}}</p>
                </div>
                <p class="generated-at">Сформировано: {{.GeneratedAt}}</p>
            </div>
        </header>

        <div class="dashboard-grid">
            <div class="sidebar">
                <section>
                    <span class="section-label">Программа / Услуга</span>
                    <div class="info-card">
                        <p class="program-name">{{.CurrentService}}</p>
                        <span class="program-sub">Текущий курс</span>
                    </div>
                </section>

                <section>
                    <span class="section-label">Документация</span>
                    <div class="docs-list">
                        {{range .Documents}}
                        <div class="attachment-item">
                            <div class="dot {{if .IsVoice}}dot-purple{{else}}dot-blue{{end}}"></div>
                            {{.Name}}
                        </div>
                        {{else}}
                        <div style="font-size: 11px; color: #94a3b8; font-style: italic;">Вложения отсутствуют.</div>
                        {{end}}
                    </div>
                </section>
            </div>

            <div class="main-content">
                <section>
                    <span class="section-label">Клинический статус и примечания</span>
                    <div class="notes-box">
                        <p class="notes-text">"{{.TherapistNotes}}"</p>
                    </div>
                </section>

                {{if .VoiceTranscripts}}
                <section>
                    <span class="section-label">Автоматические расшифровки</span>
                    <div class="transcript-box">
                        <p class="transcript-text">{{.VoiceTranscripts}}</p>
                    </div>
                </section>
                {{end}}

                <section>
                    <span class="section-label">История посещений</span>
                    <div class="table-container">
                        <table>
                            <thead>
                                <tr>
                                    <th>Дата и время</th>
                                    <th>Тип записи</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr>
                                    <td>{{.LastVisit}}</td>
                                    <td>{{.CurrentService}} (Последняя запись)</td>
                                </tr>
                                <tr>
                                    <td>{{.FirstVisit}}</td>
                                    <td>{{.CurrentService}} (Первая запись)</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </section>
            </div>
        </div>

        <footer class="footer">
            <p class="footer-text">Официальный электронный медицинский документ Vera Massage Bot</p>
        </footer>
    </div>
</body>
</html>
`
