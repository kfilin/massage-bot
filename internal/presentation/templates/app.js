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

// Back Button Management — uses sessionStorage instead of window.history
// because TWA WebView does not track full-page navigations in history.
function setupBackButton() {
    const returnTo = sessionStorage.getItem('twa_return_to');

    if (returnTo) {
        tg.BackButton.show();
        tg.BackButton.onClick(() => {
            tg.BackButton.hide();              // hide before navigating — TWA retains state across page loads
            sessionStorage.removeItem('twa_return_to');
            window.location.href = returnTo;
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

    // Wire up "Show more" pagination button(s) on the history tab
    document.body.addEventListener('click', (e) => {
        const btn = e.target.closest('.btn-show-more');
        if (!btn) return;
        e.preventDefault();
        loadMoreHistory(parseInt(btn.dataset.nextOffset, 10), parseInt(btn.dataset.limit, 10), btn);
    });
});

// History pagination: fetch the next page from the same card endpoint
// with ?partial=history, append the rendered cards to the history list,
// and replace (or remove) the show-more button.
async function loadMoreHistory(offset, limit, btn) {
    if (btn) {
        btn.disabled = true;
        btn.textContent = 'Загрузка...';
    }
    try {
        const url = new URL(window.location.href);
        url.searchParams.set('offset', String(offset));
        url.searchParams.set('limit', String(limit));
        url.searchParams.set('partial', 'history');
        const resp = await fetch(url.toString(), { credentials: 'same-origin' });
        if (!resp.ok) throw new Error('HTTP ' + resp.status);
        const html = await resp.text();
        const container = document.getElementById('content-history');
        if (!container) return;
        // The response contains visit cards followed by an optional new
        // show-more button. Parse with a temporary wrapper, then append
        // every node to the history list. Any new show-more button
        // included in the response will end up at the end of the list.
        const tmp = document.createElement('div');
        tmp.innerHTML = html;
        // Remove the old button before appending the new content
        if (btn) btn.closest('.show-more-row')?.remove();
        while (tmp.firstChild) {
            container.appendChild(tmp.firstChild);
        }
        tg.HapticFeedback.selectionChanged();
    } catch (e) {
        if (btn) {
            btn.disabled = false;
            btn.textContent = 'Показать ещё';
        }
        tg.showAlert('Не удалось загрузить ещё. Попробуйте ещё раз.');
        console.error('loadMoreHistory failed:', e);
    }
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

function openMedia(url, type) {
    window.open(url, '_blank');
}
