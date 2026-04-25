// Vera Massage Clinic — TWA Interactivity

const tg = window.Telegram.WebApp;
tg.expand();
tg.ready();

// Tab Switching Logic
function switchSegment(el) {
    if (!el) return;
    
    // UI Update
    document.querySelectorAll('.segment-item').forEach(item => item.classList.remove('active'));
    el.classList.add('active');

    // Visibility Update
    const target = el.getAttribute('data-target');
    document.querySelectorAll('.segment-content').forEach(content => {
        content.style.display = 'none';
    });
    
    const content = document.getElementById('content-' + target);
    if (content) {
        content.style.display = 'block';
        // Simple animation
        content.style.opacity = '0';
        setTimeout(() => {
            content.style.transition = 'opacity 0.3s ease';
            content.style.opacity = '1';
        }, 10);
    }
    
    // Haptic Feedback
    tg.HapticFeedback.selectionChanged();
}

// Back Button Management
function setupBackButton() {
    // If we have a 'back=true' in URL or if history > 1, show back button
    const urlParams = new URLSearchParams(window.location.search);
    const canGoBack = window.history.length > 1 || urlParams.has('from_search');

    if (canGoBack) {
        tg.BackButton.show();
        tg.BackButton.onClick(() => {
            window.history.back();
        });
    } else {
        tg.BackButton.hide();
    }
}

// Initialize on Load
window.addEventListener('DOMContentLoaded', () => {
    setupBackButton();
    
    // Set Theme Colors
    tg.setBackgroundColor('#F0F4F0');
    tg.setHeaderColor('#F0F4F0');
});

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

function openMedia(url, type) {
    window.open(url, '_blank');
}
