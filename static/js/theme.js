/**
 * Theme manager — persists light/dark preference in localStorage.
 * Must be imported before app.js initialises so the correct theme
 * is applied before first render.
 */

const STORAGE_KEY = 'study-app-theme';

function applyTheme(theme) {
  if (theme === 'light') {
    document.documentElement.setAttribute('data-theme', 'light');
  } else {
    document.documentElement.removeAttribute('data-theme');
  }
  const moon = document.getElementById('theme-icon-moon');
  const sun  = document.getElementById('theme-icon-sun');
  if (moon) moon.style.display = theme === 'light' ? 'none'  : '';
  if (sun)  sun.style.display  = theme === 'light' ? ''      : 'none';
}

export function initTheme() {
  const saved = localStorage.getItem(STORAGE_KEY) ?? 'dark';
  applyTheme(saved);

  document.getElementById('theme-btn')?.addEventListener('click', () => {
    const current = document.documentElement.getAttribute('data-theme') === 'light' ? 'light' : 'dark';
    const next    = current === 'light' ? 'dark' : 'light';
    localStorage.setItem(STORAGE_KEY, next);
    applyTheme(next);
  });
}
