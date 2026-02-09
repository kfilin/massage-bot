package storage

const medicalRecordTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no, viewport-fit=cover">
    <title>Medical Card</title>
    <script src="https://telegram.org/js/telegram-web-app.js"></script>
    <link href="https://fonts.googleapis.com/css2?family=SF+Pro+Display:wght@400;500;600;700&family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-page: #F2F2F7; /* Apple System Gray 6 */
            --bg-card: #FFFFFF;
            --text-primary: #000000;
            --text-secondary: #8E8E93;
            --accent: #007AFF; /* Apple Blue */
            --danger: #FF3B30;
            --separator: #C6C6C8;
            --glass: rgba(255, 255, 255, 0.9);
        }

        body {
            background-color: var(--bg-page);
            font-family: -apple-system, BlinkMacSystemFont, "SF Pro Display", "Inter", sans-serif;
            margin: 0;
            padding: 20px;
            color: var(--text-primary);
            -webkit-font-smoothing: antialiased;
            padding-bottom: 100px; /* Space for FAB */
        }

        /* Dark Mode Support */
        @media (prefers-color-scheme: dark) {
            :root {
                --bg-page: #000000;
                --bg-card: #1C1C1E;
                --text-primary: #FFFFFF;
                --text-secondary: #8E8E93;
                --accent: #0A84FF;
                --danger: #FF453A;
                --separator: #38383A;
                --glass: rgba(30, 30, 30, 0.9);
            }
        }

        .header {
            margin-bottom: 24px;
        }

        .large-title {
            font-size: 34px;
            font-weight: 700;
            letter-spacing: -0.02em;
            margin: 0 0 8px 0;
            color: var(--text-primary);
        }

        .subtitle {
            font-size: 13px;
            color: var(--text-secondary);
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            margin-bottom: 4px;
        }

        /* CARD STYLE (Apple Health) */
        .card {
            background: var(--bg-card);
            border-radius: 12px;
            padding: 16px;
            margin-bottom: 16px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.05);
        }

        /* Stats Grid */
        .stats-grid {
            display: grid; 
            grid-template-columns: 1fr 1fr; 
            gap: 12px; 
            margin-bottom: 30px;
        }

        .stat-label {
            font-size: 12px; 
            font-weight: 600; 
            color: var(--text-secondary); 
            text-transform: uppercase;
        }

        .stat-value {
            font-size: 18px; 
            font-weight: 700; 
            color: var(--text-primary); 
            margin-top: 4px;
        }
        .stat-value.accent { color: var(--accent); }

        .stat-sub {
            font-size: 12px; 
            color: var(--text-secondary);
            margin-top: 2px;
        }

        /* TIMELINE (Chronological) */
        .timeline-label {
            font-size: 20px;
            font-weight: 700;
            margin: 30px 0 10px 4px;
            color: var(--text-primary);
        }

        .timeline-item {
            display: flex;
            gap: 12px;
            margin-bottom: 24px;
            position: relative;
        }

        /* The vertical line */
        .timeline-item::before {
            content: '';
            position: absolute;
            left: 19px;
            top: 40px;
            bottom: -30px;
            width: 2px;
            background: var(--separator);
            z-index: 0;
            opacity: 0.3;
        }
        .timeline-item:last-child::before { display: none; }

        .timeline-icon-box {
            width: 40px;
            height: 40px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            flex-shrink: 0;
            z-index: 1;
            font-size: 20px;
            background: var(--bg-card);
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
        }
        .icon-blue { background: #E1F0FF; color: #007AFF; }
        .icon-orange { background: #FFF3E0; color: #FF9500; }
        .icon-green { background: #E0F2F1; color: #34C759; }
        .icon-red { background: #FFE5E5; color: #FF3B30; }

        @media (prefers-color-scheme: dark) {
            .icon-blue { background: #0040DD; color: white; }
            .icon-orange { background: #CC7700; color: white; }
            .icon-green { background: #007D35; color: white; }
            .icon-red { background: #AA1E19; color: white; }
        }

        .timeline-content {
            background: var(--bg-card);
            border-radius: 12px;
            padding: 14px;
            flex-grow: 1;
            box-shadow: 0 1px 2px rgba(0,0,0,0.04);
        }

        .timeline-date {
            font-size: 13px;
            color: var(--text-secondary);
            margin-bottom: 4px;
        }

        .timeline-title {
            font-size: 16px;
            font-weight: 600;
            margin-bottom: 6px;
            color: var(--text-primary);
        }

        .timeline-body {
            font-size: 15px;
            line-height: 1.5;
            color: var(--text-primary);
            white-space: pre-wrap;
        }

        .badge {
            display: inline-block;
            padding: 2px 8px;
            border-radius: 6px;
            background: var(--bg-page);
            color: var(--text-secondary);
            font-size: 12px;
            font-weight: 500;
            margin-top: 8px;
        }

        /* FAB */
        .fab {
            position: fixed;
            bottom: 24px;
            right: 24px;
            width: 56px;
            height: 56px;
            background: var(--accent);
            border-radius: 28px;
            box-shadow: 0 4px 12px rgba(0,122,255,0.3);
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
            font-size: 24px;
            cursor: pointer;
            border: none;
            transition: transform 0.2s;
            z-index: 100;
        }
        .fab:active { transform: scale(0.95); }

        /* Modals & Lightbox */
        .modal {
            display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%;
            background: rgba(0,0,0,0.5); z-index: 3000; justify-content: center; align-items: flex-end;
            backdrop-filter: blur(4px);
        }
        .modal.visible { display: flex; }
        .modal-content {
            background: var(--bg-card); width: 100%; max-width: 600px; padding: 24px;
            border-top-left-radius: 20px; border-top-right-radius: 20px;
            box-shadow: 0 -4px 20px rgba(0,0,0,0.2); animation: slideUp 0.3s ease;
            max-height: 90vh; overflow-y: auto;
        }
        @keyframes slideUp { from { transform: translateY(100%); } to { transform: translateY(0); } }
        
        .collapsible-group { margin-bottom: 24px; }
        .collapsible-header { 
            display: flex; justify-content: space-between; align-items: center; 
            font-size: 20px; font-weight: 700; color: var(--text-primary); 
            cursor: pointer; padding: 10px 4px;
        }
        .collapsible-header::after { content: '⌄'; font-size: 24px; transition: transform 0.2s; }
        .collapsible-header.collapsed::after { transform: rotate(-90deg); }
        .collapsible-content { display: block; }
        .collapsible-content.collapsed { display: none; }
        .collapsible-content.collapsed .timeline-item:nth-child(n+3) { display: none; }

        
        .form-group { margin-bottom: 16px; }
        .form-label { display: block; font-size: 13px; font-weight: 600; color: var(--text-secondary); margin-bottom: 6px; }
        .form-input { 
            width: 100%; padding: 12px; border: 1px solid var(--separator); border-radius: 10px;
            font-size: 16px; font-family: inherit; background: var(--bg-page); color: var(--text-primary);
            box-sizing: border-box;
        }
        .form-textarea { min-height: 120px; resize: vertical; }
        .btn-block { background: var(--accent); color: white; width: 100%; padding: 14px; border-radius: 12px; font-size: 16px; font-weight: 700; border: none; cursor: pointer; text-align: center; }
        .btn-block:disabled { opacity: 0.7; }
        
        .lightbox { 
            display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%; 
            background: rgba(0,0,0,0.95); z-index: 2000; justify-content: center; align-items: center; 
            flex-direction: column;
        }
        .lightbox.visible { display: flex; }
        .lightbox-content { max-width: 95%; max-height: 80vh; object-fit: contain; }
        .lightbox-close { position: absolute; top: 20px; right: 20px; color: white; font-size: 30px; background: none; border: none; cursor: pointer; }

        /* Recording UI inside Modal */
        .recording-viz {
            height: 60px; display: flex; align-items: center; justify-content: center; gap: 4px; margin: 20px 0;
        }
        .bar { width: 6px; height: 20px; background: var(--accent); border-radius: 3px; animation: wave 1s infinite; }
        .bar:nth-child(2) { animation-delay: 0.1s; }
        .bar:nth-child(3) { animation-delay: 0.2s; }
        .bar:nth-child(4) { animation-delay: 0.3s; }
        .bar:nth-child(5) { animation-delay: 0.4s; }
        @keyframes wave { 0%, 100% { height: 20px; opacity: 0.5; } 50% { height: 50px; opacity: 1; } }
        
        .recording-controls { display: flex; gap: 10px; justify-content: center; }
        .btn-record { width: 60px; height: 60px; border-radius: 30px; border: 4px solid var(--separator); background: var(--danger); cursor: pointer; display: flex; align-items: center; justify-content: center; color: white; font-size: 24px; }
        .btn-record.recording { animation: pulse-record 2s infinite; border-color: var(--danger); }
        @keyframes pulse-record { 0% { box-shadow: 0 0 0 0 rgba(255, 59, 48, 0.4); } 70% { box-shadow: 0 0 0 20px rgba(255, 59, 48, 0); } 100% { box-shadow: 0 0 0 0 rgba(255, 59, 48, 0); } }

    </style>
    <script>
        // Init Telegram
        (function() {
            const tg = window.Telegram.WebApp;
            const url = new URL(window.location.href);
            if (tg.initData && !url.searchParams.get('initData')) {
                url.searchParams.set('initData', tg.initData);
                window.location.replace(url.toString());
            }
            tg.expand();
            tg.ready();
        })();

        // --- Logic: Recording & Transcription ---
        let mediaRecorder;
        let audioChunks = [];

        async function toggleRecording() {
            const btn = document.getElementById('recordBtn');
            const status = document.getElementById('recordStatus');
            const viz = document.getElementById('recordingViz');

            if (!mediaRecorder || mediaRecorder.state === "inactive") {
                // START RECORDING
                try {
                    // Force a dummy AudioContext to "wake up" audio on iOS
                    const AudioContext = window.AudioContext || window.webkitAudioContext;
                    if (AudioContext) {
                        const ctx = new AudioContext();
                        if (ctx.state === 'suspended') ctx.resume();
                    }

                    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
                        throw new Error("Ваш браузер не поддерживает доступ к микрофону. Используйте современный браузер и HTTPS.");
                    }

                    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
                    
                    const mimeTypes = [
                        'audio/webm;codecs=opus',
                        'audio/webm',
                        'audio/mp4',
                        'audio/aac',
                        'audio/ogg'
                    ];
                    let selectedMimeType = '';
                    for (const type of mimeTypes) {
                        if (MediaRecorder.isTypeSupported(type)) {
                            selectedMimeType = type;
                            break;
                        }
                    }

                    mediaRecorder = new MediaRecorder(stream, selectedMimeType ? { mimeType: selectedMimeType } : {});
                    audioChunks = [];
                    mediaRecorder.addEventListener("dataavailable", event => {
                        audioChunks.push(event.data);
                    });
                    mediaRecorder.addEventListener("stop", async () => {
                        const audioBlob = new Blob(audioChunks, { type: mediaRecorder.mimeType || 'audio/webm' });
                        await sendAudio(audioBlob);
                    });
                    
                    mediaRecorder.start();
                    btn.classList.add('recording');
                    viz.style.display = 'flex';
                    status.innerText = "Слушаю...";
                } catch (e) {
                    alert("Ошибка микрофона: " + e.message + "\n\nУбедитесь, что вы используете HTTPS и дали разрешение в настройках браузера/Telegram.");
                    console.error(e);
                }
            } else {
                // STOP RECORDING
                mediaRecorder.stop();
                btn.classList.remove('recording');
                viz.style.display = 'none';
                status.innerText = "Обработка...";
            }
        }

        async function sendAudio(blob) {
            const tg = window.Telegram.WebApp;
            const formData = new FormData();
            formData.append('voice', blob, 'recording.mp3');
            formData.append('initData', tg.initData);

            try {
                const resp = await fetch('/api/transcribe', {
                    method: 'POST',
                    body: formData
                });
                const res = await resp.json();
                
                if (res.status === 'ok') {
                    // Append text to the editor
                    const editor = document.getElementById('editNotes');
                    const current = editor.value;
                    const timestamp = new Date().toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
                    editor.value = current + (current ? '\n\n' : '') + '**Голосовая заметка (' + timestamp + '):**\n' + res.text;
                    document.getElementById('recordStatus').innerText = "Готово!";
                } else {
                    alert("Ошибка транскрибации: " + res.error);
                    document.getElementById('recordStatus').innerText = "Ошибка.";
                }
            } catch (e) {
                console.error(e);
                alert("Сетевая ошибка при транскрибации.");
                document.getElementById('recordStatus').innerText = "Ошибка.";
            }
        }

        // --- Logic: Patient Update ---
        function openEditModal() {
            document.getElementById('editModal').classList.add('visible');
        }
        function closeEditModal() {
             document.getElementById('editModal').classList.remove('visible');
        }

        async function savePatient(e) {
            e.preventDefault();
            const btn = document.getElementById('saveBtn');
            const originalText = btn.innerText;
            btn.innerText = 'Сохранение...';
            btn.disabled = true;

            const name = document.getElementById('editName').value;
            const notes = document.getElementById('editNotes').value;
            const tg = window.Telegram.WebApp;

            try {
                const resp = await fetch('/api/patient/update', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        initData: tg.initData,
                        id: '{{.TelegramID}}',
                        name: name,
                        notes: notes
                    })
                });
                const res = await resp.json();
                if (res.status === 'ok') {
                    tg.showAlert('Сохранено.');
                    window.location.reload();
                } else {
                    tg.showAlert('Ошибка: ' + res.error);
                }
            } catch (err) {
                tg.showAlert('Ошибка сети');
            } finally {
                btn.innerText = originalText;
                btn.disabled = false;
            }
        }

        // --- Logic: Media ---
         function openMedia(url, type) {
            const lb = document.getElementById('lightbox');
            const container = document.getElementById('lightbox-media');
            container.innerHTML = '';
            
            if (type === 'image' || type === 'photo' || type === 'scan') type = 'image';
            
            try {
                if (type === 'video') {
                    const vid = document.createElement('video');
                    vid.src = url; vid.controls = true; vid.className = 'lightbox-content';
                    container.appendChild(vid);
                } else if (type === 'voice' || type === 'audio') {
                    const audio = document.createElement('audio');
                    audio.src = url; audio.controls = true; audio.className = 'lightbox-content';
                    container.appendChild(audio);
                } else if (type === 'image') {
                    const img = document.createElement('img');
                    img.src = url; img.className = 'lightbox-content';
                    container.appendChild(img);
                } else {
                     window.open(url, '_blank');
                     return;
                }
                lb.classList.add('visible');
            } catch (e) {
                console.error(e);
            }
        }
        function closeLightbox() {
             document.getElementById('lightbox').classList.remove('visible');
             document.getElementById('lightbox-media').innerHTML = '';
        }

        function cancelAppointment(event, apptId, btn) {
             event.preventDefault();
             btn.innerText = "...";
             const tg = window.Telegram.WebApp;
             fetch("/cancel", { 
                 method: "POST", headers: { "Content-Type": "application/json" },
                 body: JSON.stringify({ initData: tg.initData, apptId: apptId })
             }).then(r => r.json()).then(res => {
                 if (res.status === "ok") {
                     btn.closest('.card').remove();
                     tg.showAlert("Отменено");
                     setTimeout(() => window.location.reload(), 1000);
                 } else {
                     tg.showAlert("Ошибка: " + res.error);
                     btn.innerText = "Отменить";
                 }
             });
        }
    </script>
</head>
<body>

    <!-- Header Section -->
    <div class="header">
        <div style="display: flex; justify-content: space-between; align-items: center;">
            <div class="subtitle">МЕДИЦИНСКАЯ КАРТА</div>
            <div style="font-size: 10px; color: var(--text-secondary);">{{.BotVersion}}</div>
        </div>
        <div style="display: flex; justify-content: space-between; align-items: flex-end;">
            <h1 class="large-title">{{.Name}}</h1>
            <div onclick="openEditModal()" style="width: 38px; height: 38px; background: #E5E5EA; border-radius: 19px; display: flex; align-items: center; justify-content: center; margin-bottom: 6px; cursor: pointer;">
                👤
            </div>
        </div>
    </div>

    <!-- Stats Row -->
    <div class="stats-grid">
        <div class="card" style="margin: 0; padding: 14px;">
            <div class="stat-label">След. визит</div>
            {{if .FutureAppointments}}
                {{$next := index .FutureAppointments 0}}
                <div class="stat-value accent">{{$next.Date}}</div>
                <div class="stat-sub">{{$next.Service}}</div>
                {{if $next.CanCancel}}
                    <button onclick="cancelAppointment(event, '{{$next.ID}}', this)" style="margin-top:8px; font-size:11px; color:#FF3B30; background:none; border:1px solid #FF3B30; padding:4px 12px; border-radius:4px;">Отменить</button>
                {{end}}
            {{else}}
                <div class="stat-value" style="color:var(--text-secondary); font-size:14px;">Нет</div>
                <div class="stat-sub"><a href="https://t.me/{{.BotUsername}}?start=book" style="color:var(--accent); text-decoration:none;">Записаться</a></div>
            {{end}}
        </div>
        <div class="card" style="margin: 0; padding: 14px;">
            <div class="stat-label">Всего визитов</div>
            <div class="stat-value">{{.TotalVisits}}</div>
            <div class="stat-sub">С {{if or .FirstVisit.IsZero (lt .FirstVisit.Year 2000)}}-{{else}}{{.FirstVisit.Format "Jan 2006"}}{{end}}</div>
        </div>
    </div>

    <!-- Timeline Stream -->
    <!-- Items: Therapist Notes -->
    {{if .TherapistNotes}}
    <div class="timeline-item">
        <div class="timeline-icon-box icon-orange">📝</div>
        <div class="timeline-content">
            <div class="timeline-date">Последние записи</div>
            <div class="timeline-title">Резюме терапевта</div>
            <!-- Fixed: Ensure HTML is rendered, not as text -->
            <div class="timeline-body" style="white-space: normal;">{{.TherapistNotes}}</div>
        </div>
    </div>
    {{end}}

    <div class="collapsible-group">
        <div class="collapsible-header" onclick="this.classList.toggle('collapsed'); this.nextElementSibling.classList.toggle('collapsed');">
            Хронология визитов
        </div>
        <div class="collapsible-content">


    <!-- Item: Recent Visits -->
    {{range .RecentVisits}}
    <div class="timeline-item">
        <div class="timeline-icon-box icon-blue">👋</div>
        <div class="timeline-content">
            <div class="timeline-date">{{.Date}}</div>
            <div class="timeline-title">Визит ({{.Service}})</div>
            <div class="badge">#подтвержден</div>
        </div>
    </div>
    {{end}}
        </div>
    </div>


    <!-- Item: Documents / Media -->
    {{range .DocGroups}}
        <div class="collapsible-group">
            <div class="collapsible-header collapsed" onclick="this.classList.toggle('collapsed'); this.nextElementSibling.classList.toggle('collapsed');">
                {{.Name}} ({{.Count}})
            </div>
            <div class="collapsible-content collapsed">
                {{range .Files}}
                <div class="timeline-item">
                    <div class="timeline-icon-box icon-green">
                        {{if eq .FileType "voice"}}🎙️{{else if eq .FileType "image"}}🖼️{{else}}📄{{end}}
                    </div>
                    <div class="timeline-content">
                        <div class="timeline-date">{{.CreatedAt.Format "02.01 15:04"}}</div>
                        {{if eq .FileType "voice"}}
                             <button onclick="openMedia('/api/media/{{.ID}}', 'voice')" style="margin-top:8px; background:var(--bg-page); border:none; padding:6px 12px; border-radius:8px;">▶ Прослушать</button>
                        {{else}}
                             <button onclick="openMedia('/api/media/{{.ID}}', '{{.FileType}}')" style="color: var(--accent); background: none; border: none; font-weight: 600; padding: 0; margin-top:4px;">Открыть файл</button>
                        {{end}}
                    </div>
                </div>
                {{end}}
            </div>
        </div>
    {{end}}

    
    <!-- Floating Action Button -->
    <button class="fab" onclick="openEditModal()" title="Add Note / Voice">🎤</button>
    <!-- We open Edit Modal which now has Voice capabilities! -->

    <!-- Edit/Voice Modal -->
    <div id="editModal" class="modal" onclick="if(event.target === this) closeEditModal()">
        <div class="modal-content">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
                <h2 style="margin: 0; font-size: 20px; color: var(--text-primary);">Редактор и Голос</h2>
                <button onclick="closeEditModal()" style="background:none; border:none; font-size: 24px; color: var(--text-secondary); cursor:pointer;">✕</button>
            </div>

            <!-- Voice Controls -->
            <div style="border-bottom: 1px solid var(--separator); padding-bottom: 20px; margin-bottom: 20px; text-align: center;">
                <div class="recording-controls">
                    <button id="recordBtn" class="btn-record" onclick="toggleRecording()">🎙️</button>
                </div>
                <div id="recordingViz" class="recording-viz" style="display:none;">
                    <div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div>
                </div>
                <div id="recordStatus" style="margin-top:10px; font-size:13px; color:var(--text-secondary);">Нажмите для записи</div>
            </div>

            <form onsubmit="savePatient(event)">
                <div class="form-group">
                    <label class="form-label">Имя</label>
                    <input type="text" id="editName" class="form-input" value="{{.Name}}" required>
                </div>
                <div class="form-group">
                    <label class="form-label">Заметки (Markdown поддерживается)</label>
                    <textarea id="editNotes" class="form-input form-textarea">{{.RawNotes}}</textarea>
                </div>
                <button type="submit" id="saveBtn" class="btn-block">Сохранить изменения</button>
            </form>
        </div>
    </div>

    <!-- Lightbox -->
    <div id="lightbox" class="lightbox" onclick="if(event.target === this) closeLightbox()">
        <button class="lightbox-close" onclick="closeLightbox()">✕</button>
        <div id="lightbox-media"></div>
    </div>

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
