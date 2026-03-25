/**
 * Desktop integration — always-on-top pin toggle.
 * goToggleAlwaysOnTop is bound by the Go webview2 host.
 * Falls back silently when running in a plain browser (no binding).
 */

const PIN_KEY = 'study-app-pinned';

export function initDesktop() {
  const btn = document.getElementById('pin-btn');
  if (!btn) return;

  // Restore persisted state
  const pinned = localStorage.getItem(PIN_KEY) === 'true';
  if (pinned) {
    _setPinned(btn, true);
    _callGoToggle(); // sync Go-side state on startup
  }

  btn.addEventListener('click', () => {
    const next = btn.classList.contains('is-unpinned');
    _setPinned(btn, next);
    localStorage.setItem(PIN_KEY, String(next));
    _callGoToggle();
  });
}

function _setPinned(btn, pinned) {
  btn.classList.toggle('is-pinned',   pinned);
  btn.classList.toggle('is-unpinned', !pinned);
  btn.title = pinned ? 'Unpin window' : 'Pin window on top';
  btn.setAttribute('aria-pressed', String(pinned));
}

function _callGoToggle() {
  if (typeof window.goToggleAlwaysOnTop === 'function') {
    window.goToggleAlwaysOnTop().catch(() => {});
  }
}
