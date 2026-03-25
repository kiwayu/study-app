/**
 * Day ruler — faint time-of-day context strip fixed to the right edge.
 * Renders hour marks for 06:00–23:00 and a live "now" indicator.
 * Updates every 60 seconds.
 */

const DAY_START_H = 6;   // 06:00
const DAY_END_H   = 23;  // 23:00
const DAY_SPAN_H  = DAY_END_H - DAY_START_H; // 17 hours

export function initRuler() {
  const el = document.getElementById('day-ruler');
  if (!el) return;

  _render(el);
  // Update every 60 seconds to move the "now" indicator
  setInterval(() => _render(el), 60_000);
}

function _render(el) {
  const now   = new Date();
  const nowH  = now.getHours() + now.getMinutes() / 60;
  const nowPct = Math.max(0, Math.min(100, (nowH - DAY_START_H) / DAY_SPAN_H * 100));

  // Build hour tick marks
  let html = '';
  for (let h = DAY_START_H; h <= DAY_END_H; h++) {
    const pct = (h - DAY_START_H) / DAY_SPAN_H * 100;
    const label = `${String(h).padStart(2, '0')}:00`;
    html += `<div class="ruler-tick" style="top:${pct.toFixed(2)}%">${label}</div>`;
  }

  // Current time indicator
  const timeStr = `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}`;
  html += `<div class="ruler-now" style="top:${nowPct.toFixed(2)}%">
    <span class="ruler-now-dot"></span>
    <span class="ruler-now-time">${timeStr}</span>
  </div>`;

  el.innerHTML = html;
}
