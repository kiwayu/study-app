import { getCompletions } from './api.js';

/**
 * StatsPanel — GitHub-style task completion heatmap.
 * Shows the last 12 months (year view) or current month as a day grid.
 */
export class StatsPanel {
  constructor(panelEl, overlayEl) {
    this._panel   = panelEl;
    this._overlay = overlayEl;
    this._view    = 'year'; // 'year' | 'month'

    this._panel.querySelector('[data-stats-close]')
      ?.addEventListener('click', () => this.close());
    this._overlay.addEventListener('click', () => this.close());

    document.addEventListener('keydown', e => {
      if (e.key === 'Escape' && this._panel.classList.contains('is-open')) this.close();
    });

    this._panel.querySelector('#stats-view-year')
      ?.addEventListener('click', () => { this._view = 'year';  this._load(); });
    this._panel.querySelector('#stats-view-month')
      ?.addEventListener('click', () => { this._view = 'month'; this._load(); });
  }

  async open() {
    this._panel.classList.add('is-open');
    this._overlay.classList.add('is-open');
    await this._load();
  }

  close() {
    this._panel.classList.remove('is-open');
    this._overlay.classList.remove('is-open');
  }

  async _load() {
    const grid = this._panel.querySelector('#stats-heatmap');
    if (!grid) return;
    grid.innerHTML = '<span class="stats-loading">Loading…</span>';

    const { data, error } = await getCompletions();
    if (error || !data) {
      grid.innerHTML = '<span class="stats-loading">Could not load data.</span>';
      return;
    }

    grid.innerHTML = this._view === 'year'
      ? _buildYear(data)
      : _buildMonth(data);

    // Sync toggle button states
    this._panel.querySelector('#stats-view-year')
      ?.classList.toggle('is-active', this._view === 'year');
    this._panel.querySelector('#stats-view-month')
      ?.classList.toggle('is-active', this._view === 'month');
  }
}

// ── Builders ────────────────────────────────────────────────

/** Render a 52-week contribution grid (year view). */
function _buildYear(counts) {
  const today   = new Date();
  today.setHours(0, 0, 0, 0);
  const max = Math.max(1, ...Object.values(counts));

  // Start from the Sunday 52 weeks ago
  const start = new Date(today);
  start.setDate(start.getDate() - 364 - start.getDay());

  // Month labels
  const months = [];
  let cur = new Date(start);
  let lastMonth = -1;
  const totalWeeks = Math.ceil(365 / 7) + 1;

  for (let w = 0; w < totalWeeks; w++) {
    const m = cur.getMonth();
    if (m !== lastMonth) {
      months.push({ week: w, label: cur.toLocaleString('default', { month: 'short' }) });
      lastMonth = m;
    }
    cur.setDate(cur.getDate() + 7);
  }

  const monthBar = months.map(({ week, label }) =>
    `<span class="hm-month-label" style="grid-column:${week + 1}">${label}</span>`
  ).join('');

  // Day cells
  let cells = '';
  const d = new Date(start);
  for (let w = 0; w < totalWeeks; w++) {
    for (let day = 0; day < 7; day++) {
      const key   = _fmt(d);
      const count = counts[key] ?? 0;
      const level = count === 0 ? 0 : Math.ceil((count / max) * 4);
      const isPast = d <= today;
      const title = `${key}: ${count} task${count !== 1 ? 's' : ''} completed`;
      cells += isPast
        ? `<span class="hm-cell level-${level}" title="${title}" aria-label="${title}"></span>`
        : `<span class="hm-cell level-future"></span>`;
      d.setDate(d.getDate() + 1);
    }
  }

  return `
    <div class="hm-year-wrap">
      <div class="hm-month-row">${monthBar}</div>
      <div class="hm-day-labels">
        <span></span><span>Mon</span><span></span><span>Wed</span><span></span><span>Fri</span><span></span>
      </div>
      <div class="hm-grid">${cells}</div>
    </div>
    <div class="hm-legend">
      <span class="hm-legend-label">Less</span>
      <span class="hm-cell level-0"></span>
      <span class="hm-cell level-1"></span>
      <span class="hm-cell level-2"></span>
      <span class="hm-cell level-3"></span>
      <span class="hm-cell level-4"></span>
      <span class="hm-legend-label">More</span>
    </div>`;
}

/** Render current month as a calendar grid (month view). */
function _buildMonth(counts) {
  const today   = new Date();
  today.setHours(0, 0, 0, 0);
  const year  = today.getFullYear();
  const month = today.getMonth();
  const max   = Math.max(1, ...Object.values(counts));

  const firstDay  = new Date(year, month, 1).getDay(); // 0=Sun
  const daysInMonth = new Date(year, month + 1, 0).getDate();
  const monthName = today.toLocaleString('default', { month: 'long', year: 'numeric' });

  const dayHeaders = ['Sun','Mon','Tue','Wed','Thu','Fri','Sat']
    .map(d => `<span class="hm-cal-header">${d}</span>`).join('');

  let cells = Array(firstDay).fill('<span class="hm-cal-empty"></span>').join('');
  for (let d = 1; d <= daysInMonth; d++) {
    const date  = new Date(year, month, d);
    const key   = _fmt(date);
    const count = counts[key] ?? 0;
    const level = count === 0 ? 0 : Math.ceil((count / max) * 4);
    const isToday = date.getTime() === today.getTime();
    const title = `${key}: ${count} task${count !== 1 ? 's' : ''} completed`;
    cells += `<span class="hm-cal-cell level-${level}${isToday ? ' is-today' : ''}" title="${title}">
      <span class="hm-cal-day">${d}</span>
    </span>`;
  }

  return `
    <div class="hm-month-title">${monthName}</div>
    <div class="hm-cal-grid">
      ${dayHeaders}
      ${cells}
    </div>`;
}

function _fmt(d) {
  return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`;
}
