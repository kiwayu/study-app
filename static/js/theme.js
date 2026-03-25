/**
 * Theme manager — named themes with localStorage persistence.
 * Exported THEMES array is used by the settings drawer picker.
 */

export const THEMES = [
  // id, display name, isDark, [bg, surface, accent] preview colors
  { id: 'serika-dark',  name: 'Serika Dark',  dark: true,  c: ['#323437', '#2c2e31', '#e2b714'] },
  { id: 'serika-light', name: 'Serika Light', dark: false, c: ['#e1ded7', '#cac8be', '#e2b714'] },
  { id: 'base',         name: 'Base',         dark: false, c: ['#ffffff', '#f5f5f5', '#888888'] },
  { id: 'monokai',      name: 'Monokai',      dark: true,  c: ['#272822', '#1e1f1c', '#a6e22e'] },
  { id: 'dracula',      name: 'Dracula',      dark: true,  c: ['#282a36', '#21222c', '#bd93f9'] },
  { id: 'nord',         name: 'Nord',         dark: true,  c: ['#2e3440', '#252a33', '#88c0d0'] },
  { id: 'gruvbox-dark', name: 'Gruvbox Dark', dark: true,  c: ['#282828', '#1d2021', '#b8bb26'] },
  { id: 'terror-below', name: 'Terror Below', dark: true,  c: ['#0d0d0d', '#111111', '#c0392b'] },
  { id: 'alduin',       name: 'Alduin',       dark: true,  c: ['#1c1c1c', '#141414', '#d65d0e'] },
  { id: 'night-runner', name: 'Night Runner', dark: true,  c: ['#1a1a2e', '#16213e', '#00f5d4'] },
  { id: 'dark-plus',    name: 'Dark+',        dark: true,  c: ['#1e1e1e', '#252526', '#569cd6'] },
  { id: 'future-funk',  name: 'Future Funk',  dark: true,  c: ['#241734', '#1d1028', '#ff79c6'] },
  { id: 'laser',        name: 'Laser',        dark: true,  c: ['#1a0a2e', '#150822', '#ff00ff'] },
  { id: 'superuser',    name: 'Superuser',    dark: true,  c: ['#0f1117', '#0a0d12', '#5af78e'] },
  { id: 'cyberspace',   name: 'Cyberspace',   dark: true,  c: ['#0d1117', '#0a0d12', '#00ffff'] },
  { id: 'synthwave',    name: 'Synthwave',    dark: true,  c: ['#2b213a', '#241b31', '#f97e72'] },
  { id: 'taro',         name: 'Taro',         dark: true,  c: ['#2d2640', '#231e35', '#c4aaf0'] },
  { id: 'botanical',    name: 'Botanical',    dark: false, c: ['#e8f5e9', '#ddefde', '#4caf50'] },
  { id: 'froyo',        name: 'Froyo',        dark: false, c: ['#fce4ec', '#f5d0dd', '#e91e63'] },
  { id: 'sewing-tin',   name: 'Sewing Tin',   dark: false, c: ['#f5f0e8', '#ece7df', '#8b6914'] },
  { id: 'honey',        name: 'Honey',        dark: false, c: ['#fff8e1', '#fff0c0', '#ff8f00'] },
  { id: 'aurora',       name: 'Aurora',       dark: true,  c: ['#0d1b2a', '#0a1520', '#4dd0e1'] },
];

const STORAGE_KEY      = 'study-app-theme';
const LAST_DARK_KEY    = 'study-app-last-dark';
const LAST_LIGHT_KEY   = 'study-app-last-light';
const DEFAULT_THEME_ID = 'serika-dark';

/** Apply a theme by ID and persist the choice. */
export function setTheme(id) {
  const theme = THEMES.find(t => t.id === id) ?? THEMES[0];
  document.documentElement.setAttribute('data-theme', theme.id);
  localStorage.setItem(STORAGE_KEY, theme.id);

  // Track last used dark/light separately for the toggle button
  if (theme.dark) {
    localStorage.setItem(LAST_DARK_KEY, theme.id);
  } else {
    localStorage.setItem(LAST_LIGHT_KEY, theme.id);
  }

  _syncMoonIcon(theme.dark);
  _syncPickerActive(theme.id);
}

/** Populate a container element with theme swatch buttons. */
export function buildThemePicker(containerEl) {
  containerEl.innerHTML = '';
  THEMES.forEach(theme => {
    const btn = document.createElement('button');
    btn.className = 'theme-swatch';
    btn.dataset.themeId = theme.id;
    btn.title = theme.name;
    btn.setAttribute('aria-label', `Apply ${theme.name} theme`);
    btn.innerHTML =
      `<span class="swatch-dot" style="background:${theme.c[0]}"></span>` +
      `<span class="swatch-dot" style="background:${theme.c[1]}"></span>` +
      `<span class="swatch-dot" style="background:${theme.c[2]}"></span>` +
      `<span class="swatch-name">${theme.name}</span>`;
    btn.addEventListener('click', () => setTheme(theme.id));
    containerEl.appendChild(btn);
  });

  const current = document.documentElement.getAttribute('data-theme') ?? DEFAULT_THEME_ID;
  _syncPickerActive(current);
}

/** Wire the moon/sun toggle button (quick dark↔light switch). */
export function initTheme() {
  const saved = localStorage.getItem(STORAGE_KEY) ?? DEFAULT_THEME_ID;
  const theme = THEMES.find(t => t.id === saved) ?? THEMES[0];
  document.documentElement.setAttribute('data-theme', theme.id);
  _syncMoonIcon(theme.dark);

  document.getElementById('theme-btn')?.addEventListener('click', () => {
    const currentId  = document.documentElement.getAttribute('data-theme') ?? DEFAULT_THEME_ID;
    const current    = THEMES.find(t => t.id === currentId) ?? THEMES[0];
    const nextId     = current.dark
      ? (localStorage.getItem(LAST_LIGHT_KEY) ?? 'serika-light')
      : (localStorage.getItem(LAST_DARK_KEY)  ?? DEFAULT_THEME_ID);
    setTheme(nextId);
  });
}

// ── Private ──────────────────────────────────────────────────

function _syncMoonIcon(isDark) {
  const moon = document.getElementById('theme-icon-moon');
  const sun  = document.getElementById('theme-icon-sun');
  if (moon) moon.style.display = isDark ? '' : 'none';
  if (sun)  sun.style.display  = isDark ? 'none' : '';
}

function _syncPickerActive(activeId) {
  document.querySelectorAll('.theme-swatch').forEach(btn => {
    btn.classList.toggle('is-active', btn.dataset.themeId === activeId);
  });
}
