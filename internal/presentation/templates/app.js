// Vera Massage Clinic — TWA Interactivity

const tg = window.Telegram.WebApp;
tg.expand();
tg.ready();

function switchSegment(el) {
    // UI Update
    document.querySelectorAll('.segment-item').forEach(item => item.classList.remove('active'));
    el.classList.add('active');

    // Visibility Update
    const target = el.getAttribute('data-target');
    document.querySelectorAll('.segment-content').forEach(content => {
        content.style.display = 'none';
    });
    document.getElementById('content-' + target).style.display = 'block';
    
    // Haptic Feedback
    tg.HapticFeedback.selectionChanged();
}

async function approveDraft(id) {
    if (!confirm('Добавить эту расшифровку в постоянную карту?')) return;
    
    try {
        const resp = await fetch('/api/draft/approve', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id: id, initData: tg.initData })
        });
        const res = await resp.json();
        if (res.status === 'ok') {
            tg.showAlert('Запись добавлена.');
            document.getElementById('draft-' + id).remove();
        } else {
            tg.showAlert('Ошибка: ' + res.error);
        }
    } catch (e) {
        tg.showAlert('Ошибка сети');
    }
}

async function discardDraft(id) {
    if (!confirm('Удалить этот черновик?')) return;

    try {
        const resp = await fetch('/api/draft/discard', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id: id, initData: tg.initData })
        });
        const res = await resp.json();
        if (res.status === 'ok') {
            document.getElementById('draft-' + id).remove();
        } else {
            tg.showAlert('Ошибка: ' + res.error);
        }
    } catch (e) {
        tg.showAlert('Ошибка сети');
    }
}

// Media Handling
function openMedia(url, type) {
    // Use Telegram's native photo viewer if it's an image
    if (type === 'image' || type === 'photo') {
        tg.showScanQrPopup({ text: 'Previewing image...' }); // Not ideal, but Telegram doesn't have a native image gallery API yet for local URLs
        // For now, we open in a lightbox (implemented in next step or use existing one)
    }
    window.open(url, '_blank');
}
