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
            if (tg.themeParams.bg_color) {
                document.body.style.backgroundColor = tg.themeParams.secondary_bg_color || "#f8fafc";
                document.body.style.color = tg.themeParams.text_color;
            }
        });
    </script>
    <style>
        :root {
            --primary: #2563eb;
            --primary-light: #eff6ff;
            --primary-dark: #1e40af;
            --slate-50: #f8fafc;
            --slate-100: #f1f5f9;
            --slate-200: #e2e8f0;
            --slate-300: #cbd5e1;
            --slate-500: #64748b;
            --slate-700: #334155;
            --slate-800: #1e293b;
            --slate-900: #0f172a;
        }

        * { box-sizing: border-box; -webkit-print-color-adjust: exact; }
        
        body { 
            background-color: var(--slate-100); 
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            margin: 0;
            padding: 1rem;
            line-height: 1.6;
            color: var(--slate-800);
        }

        @media (min-width: 768px) {
            body { padding: 2rem; }
        }

        .page {
            max-width: 900px;
            margin: 0 auto;
            background: white;
            padding: 20px;
            border-radius: 12px;
            box-shadow: 0 4px 20px rgba(0,0,0,0.08);
        }

        @media (min-width: 768px) {
            .page { padding: 40px; }
        }

        /* Header Section */
        .header {
            display: flex;
            flex-direction: column;
            gap: 1.5rem;
            padding-bottom: 2rem;
            margin-bottom: 2.5rem;
            border-bottom: 3px solid var(--primary);
        }

        @media (min-width: 600px) {
            .header {
                flex-direction: row;
                justify-content: space-between;
                align-items: flex-end;
            }
        }

        .header-info {
            display: flex;
            flex-direction: column;
            gap: 0.25rem;
        }

        .patient-name {
            font-size: 2.5rem;
            font-weight: 800;
            color: var(--slate-900);
            line-height: 1;
            margin: 0;
            letter-spacing: -0.02em;
        }

        .medical-id {
            font-size: 0.9rem;
            color: var(--primary);
            font-weight: 700;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        .system-tag {
            font-size: 0.75rem;
            color: var(--slate-500);
        }

        .header-stats {
            text-align: right;
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
        }

        .stat-badge {
            background: var(--primary);
            color: white;
            padding: 12px 24px;
            border-radius: 10px;
            text-align: center;
        }

        .stat-label { font-size: 0.7rem; text-transform: uppercase; font-weight: 800; opacity: 0.9; margin: 0; }
        .stat-value { font-size: 1.75rem; font-weight: 900; line-height: 1.1; margin: 0; }
        .generated-date { font-size: 0.7rem; color: var(--slate-500); font-weight: 600; text-transform: uppercase; }

        /* Content Grid */
        .main-grid {
            display: flex;
            flex-direction: column;
            gap: 2.5rem;
        }

        @media (min-width: 800px) {
            .main-grid {
                flex-direction: row;
            }
        }

        .col-sidebar {
            flex: 0 0 300px;
            display: flex;
            flex-direction: column;
            gap: 2rem;
        }

        .col-main {
            flex: 1;
            display: flex;
            flex-direction: column;
            gap: 2.5rem;
        }

        /* Components */
        .section-title {
            font-size: 0.75rem;
            font-weight: 800;
            color: var(--slate-500);
            text-transform: uppercase;
            letter-spacing: 0.12em;
            margin-bottom: 1rem;
            display: block;
        }

        .card {
            background: var(--slate-50);
            border: 1px solid var(--slate-200);
            border-radius: 10px;
            padding: 1.5rem;
            transition: all 0.2s ease;
        }

        .service-name { font-size: 1.25rem; font-weight: 700; color: var(--slate-900); margin: 0; }
        .service-status { font-size: 0.75rem; color: var(--primary); font-weight: 700; text-transform: uppercase; margin-top: 4px; }

        .note-box {
            background: var(--primary-light);
            border: 1px solid #bfdbfe;
            border-left: 5px solid var(--primary);
            padding: 1.5rem;
            border-radius: 4px 10px 10px 4px;
        }

        .note-text { font-size: 1rem; color: var(--slate-800); white-space: pre-wrap; margin: 0; }

        .transcript-box {
            background: var(--slate-50);
            border-left: 5px solid var(--slate-500);
            padding: 1.5rem;
            border-radius: 4px 10px 10px 4px;
        }

        .transcript-text { font-size: 0.9rem; color: var(--slate-700); font-style: italic; white-space: pre-wrap; margin: 0; }

        /* Document List */
        .doc-item {
            display: flex;
            align-items: center;
            gap: 12px;
            padding: 12px;
            background: white;
            border: 1px solid var(--slate-200);
            border-radius: 8px;
            margin-bottom: 10px;
            font-size: 0.85rem;
            color: var(--slate-700);
            text-decoration: none;
        }

        .doc-icon { width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0; }
        .icon-voice { background: #a855f7; }
        .icon-file { background: var(--primary); }

        /* Ledger Table */
        .ledger { width: 100%; display: flex; flex-direction: column; border: 1px solid var(--slate-200); border-radius: 10px; overflow: hidden; }
        .ledger-row { display: flex; padding: 14px 18px; border-bottom: 1px solid var(--slate-100); }
        .ledger-row:last-child { border-bottom: none; }
        .ledger-header { background: var(--slate-100); font-size: 0.7rem; font-weight: 800; color: var(--slate-500); text-transform: uppercase; }
        
        .ledger-date { flex: 0 0 160px; font-weight: 600; }
        .ledger-desc { flex: 1; }

        /* Footer */
        .footer {
            margin-top: 4rem;
            padding-top: 2rem;
            border-top: 1px solid var(--slate-100);
            text-align: center;
        }

        .footer-text {
            font-size: 0.7rem;
            color: var(--slate-500);
            font-weight: 700;
            text-transform: uppercase;
            letter-spacing: 0.1em;
        }

        /* Print Settings */
        @media print {
            body { background: white; padding: 0; }
            .page { box-shadow: none; border: none; max-width: 100%; padding: 0; }
            .stat-badge { border: 2px solid var(--primary); color: var(--primary) !important; background: transparent !important; }
            .stat-value, .stat-label { color: var(--primary) !important; }
            .note-box { background: #f0f7ff !important; }
        }
    </style>
</head>
<body>
    <div class="page">
        <header class="header">
            <div class="header-info">
                <h1 class="patient-name">{{.Name}}</h1>
                <p class="medical-id">Медицинская карта ID {{.TelegramID}}</p>
                <p class="system-tag">Система: Vera Massage Bot</p>
            </div>
            <div class="header-stats">
                <div class="stat-badge">
                    <p class="stat-label">Посещений</p>
                    <p class="stat-value">{{.TotalVisits}}</p>
                </div>
                <p class="generated-date">Сформировано: {{.GeneratedAt}}</p>
            </div>
        </header>

        <div class="main-grid">
            <div class="col-sidebar">
                <section>
                    <span class="section-title">Программа / Услуга</span>
                    <div class="card">
                        <p class="service-name">{{.CurrentService}}</p>
                        <p class="service-status">Текущий курс</p>
                    </div>
                </section>

                <section>
                    <span class="section-title">Документация</span>
                    <div class="docs-container">
                        {{range .Documents}}
                        <div class="doc-item">
                            <div class="doc-icon {{if .IsVoice}}icon-voice{{else}}icon-file{{end}}"></div>
                            <span>{{.Name}}</span>
                        </div>
                        {{else}}
                        <p style="font-size: 0.85rem; color: var(--slate-500); font-style: italic;">Вложения отсутствуют.</p>
                        {{end}}
                    </div>
                </section>
            </div>

            <div class="col-main">
                <section>
                    <span class="section-title">Клинический статус и примечания</span>
                    <div class="note-box">
                        <p class="note-text">{{.TherapistNotes}}</p>
                    </div>
                </section>

                {{if .VoiceTranscripts}}
                <section>
                    <span class="section-title">Автоматические расшифровки</span>
                    <div class="transcript-box">
                        <div class="transcript-text">{{.VoiceTranscripts}}</div>
                    </div>
                </section>
                {{end}}

                <section>
                    <span class="section-title">История посещений</span>
                    <div class="ledger">
                        <div class="ledger-row ledger-header">
                            <div class="ledger-date">Дата и время</div>
                            <div class="ledger-desc">Описание услуги</div>
                        </div>
                        <div class="ledger-row">
                            <div class="ledger-date">{{.LastVisit}}</div>
                            <div class="ledger-desc">{{.CurrentService}} (Последний визит)</div>
                        </div>
                        <div class="ledger-row">
                            <div class="ledger-date">{{.FirstVisit}}</div>
                            <div class="ledger-desc">{{.CurrentService}} (Первый визит)</div>
                        </div>
                    </div>
                </section>
            </div>
        </div>

        <footer class="footer">
            <p class="footer-text">Официальный медицинский документ Vera Massage Bot</p>
        </footer>
    </div>
</body>
</html>
`
