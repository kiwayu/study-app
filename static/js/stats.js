import { getCompletions, getEstimationStats } from './api.js';

/**
 * StatsPanel — GitHub-style task completion heatmap.
 * Shows the last 12 months (year view) or current month as a day grid.
 */
export class StatsPanel {
  constructor(panelEl, overlayEl) {
    this._panel   = panelEl;
    this._overlay = overlayEl;
    this._view    = 'year'; // 'year' | 'month' | 'est'

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
    this._panel.querySelector('#stats-view-est')
      ?.addEventListener('click', () => { this._view = 'est';   this._load(); });
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

    // Sync toggle button states
    this._panel.querySelector('#stats-view-year')
      ?.classList.toggle('is-active', this._view === 'year');
    this._panel.querySelector('#stats-view-month')
      ?.classList.toggle('is-active', this._view === 'month');
    this._panel.querySelector('#stats-view-est')
      ?.classList.toggle('is-active', this._view === 'est');

    if (this._view === 'est') {
      const { data, error } = await getEstimationStats();
      if (error || !data) {
        grid.innerHTML = '<span class="stats-loading">Could not load data.</span>';
        return;
      }
      grid.innerHTML = _buildEstimation(data);
      return;
    }

    const { data, error } = await getCompletions();
    if (error || !data) {
      grid.innerHTML = '<span class="stats-loading">Could not load data.</span>';
      return;
    }

    grid.innerHTML = this._view === 'year'
      ? _buildYear(data)
      : _buildMonth(data);
  }
}

// ── Builders ────────────────────────────────────────────────

// Cell size constants — must match CSS .hm-cell width/height and .hm-grid gap
const HM_CELL = 13; // px
const HM_GAP  = 2;  // px
const HM_COL  = HM_CELL + HM_GAP; // px per column (15)
const HM_DAY_LABEL_W = 24; // px — matches .hm-day-labels width in CSS

/** Render a 52-week contribution grid (year view). */
function _buildYear(counts) {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const max = Math.max(1, ...Object.values(counts));

  // Start from the Sunday ~52 weeks ago
  const start = new Date(today);
  start.setDate(start.getDate() - 364 - start.getDay());

  const totalWeeks = Math.ceil(365 / 7) + 1;

  // Build month labels with pixel widths proportional to weeks spanned
  const months = [];
  let cur = new Date(start);
  let lastMonth = -1;
  for (let w = 0; w < totalWeeks; w++) {
    const m = cur.getMonth();
    if (m !== lastMonth) {
      if (months.length > 0) months[months.length - 1].weeks = w - months[months.length - 1].startWeek;
      months.push({ startWeek: w, label: cur.toLocaleString('default', { month: 'short' }), weeks: 0 });
      lastMonth = m;
    }
    cur.setDate(cur.getDate() + 7);
  }
  if (months.length > 0) months[months.length - 1].weeks = totalWeeks - months[months.length - 1].startWeek;

  // Day spacer width = day-labels width + gap between day-labels and grid
  const spacerW = HM_DAY_LABEL_W + HM_GAP * 2;
  const monthStrip = months.map(m =>
    `<span style="flex:0 0 ${m.weeks * HM_COL}px;overflow:hidden;white-space:nowrap">${m.label}</span>`
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
    <div class="hm-outer">
      <div class="hm-top-row">
        <div style="flex:0 0 ${spacerW}px"></div>
        <div class="hm-month-strip">${monthStrip}</div>
      </div>
      <div class="hm-grid-row">
        <div class="hm-day-labels">
          <span></span><span>Mon</span><span></span><span>Wed</span><span></span><span>Fri</span><span></span>
        </div>
        <div class="hm-grid">${cells}</div>
      </div>
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

/** Render estimation accuracy list. */
function _buildEstimation(tasks) {
  if (!tasks || tasks.length === 0) {
    return '<p class="task-empty-filter">No completed tasks yet.</p>';
  }

  const maxVal = Math.max(1, ...tasks.map(t => Math.max(t.estimated, t.actual)));

  const items = tasks.map(t => {
    const estPct  = Math.round((t.estimated / maxVal) * 100);
    const actPct  = Math.round((t.actual    / maxVal) * 100);
    const over    = t.actual - t.estimated;
    const badge   = over <= 0
      ? `<span class="est-badge on-track">✓ on track</span>`
      : `<span class="est-badge over">+${over} over</span>`;

    return `
      <div class="est-item">
        <div class="est-title" title="${_escAttr(t.title)}">${_escHtml(t.title)}</div>
        <div class="est-bars">
          <div class="est-bar-wrap">
            <div class="est-bar estimated" style="width:${estPct}%"></div>
            <div class="est-bar actual"    style="width:${actPct}%"></div>
          </div>
        </div>
        <div class="est-meta">
          <span>Est ${t.estimated} · Got ${t.actual}</span>
          ${badge}
        </div>
      </div>`;
  }).join('');

  return `<div class="est-list">${items}</div>`;
}

function _escHtml(s) {
  return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

function _escAttr(s) {
  return s.replace(/&/g,'&amp;').replace(/"/g,'&quot;');
}
