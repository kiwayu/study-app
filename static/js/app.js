import { getTasks, getSettings, getSession, startSession, pauseSession, stopSession, updateTotals } from './api.js';
import { StatsPanel } from './stats.js';
import { TimerEngine, SEGMENT_SEQUENCE } from './timer.js';
import { TaskManager } from './tasks.js';
import { SettingsManager } from './settings.js';
import { initTheme, buildThemePicker } from './theme.js';
import { initDesktop } from './desktop.js';
import { initRuler } from './ruler.js';

// ── Constants ────────────────────────────────────────────────

const CIRCUMFERENCE = 339.292;
const SEGMENT_LABELS = {
  focus: 'Focus', short_break: 'Short Break', long_break: 'Long Break',
};

// ── DOM refs ─────────────────────────────────────────────────

const $ = id => document.getElementById(id);
const timerDigits   = $('timer-digits');
const ringProgress  = $('timer-ring-progress');
const segmentLabel  = $('segment-label');
const pomodoroCount = $('pomodoro-count');
const startBtn      = $('start-btn');
const pauseBtn      = $('pause-btn');
const stopBtn       = $('stop-btn');
const settingsBtn   = $('settings-btn');
const sortBtn       = $('sort-btn');
const statsBtn      = $('stats-btn');
const addTaskForm   = $('add-task-form');
const toastContainer = $('toast-container');

// ── State ────────────────────────────────────────────────────

let currentSession  = null;
let currentSettings = null;
let lastWaterAt     = 0;
let lastStretchAt   = 0;
let _segmentTransitioning = false;

// ── Modules ──────────────────────────────────────────────────

const timer      = new TimerEngine();
const tasks      = new TaskManager($('task-list'));
const settings   = new SettingsManager($('settings-drawer'), $('settings-overlay'));
const statsPanel = new StatsPanel($('stats-panel'), $('stats-overlay'));

// ── Timer display ────────────────────────────────────────────

function fmtTime(ms) {
  const s = Math.max(0, Math.ceil(ms / 1000));
  return `${String(Math.floor(s / 60)).padStart(2, '0')}:${String(s % 60).padStart(2, '0')}`;
}

function renderTick(remainingMs, progress) {
  timerDigits.textContent = fmtTime(remainingMs);
  ringProgress.style.strokeDashoffset = CIRCUMFERENCE * (1 - progress);
}

function setSegmentLabel(type) {
  segmentLabel.textContent = SEGMENT_LABELS[type] ?? type;
}

function setPomodoroCount(n) {
  pomodoroCount.textContent = `${n} / 4`;
}

function triggerSegmentPulse() {
  segmentLabel.classList.remove('segment-pulse');
  void segmentLabel.offsetWidth; // force reflow to restart animation
  segmentLabel.classList.add('segment-pulse');
}

function setControlState(status) {
  if (status === 'running') {
    startBtn.style.display = 'none';
    pauseBtn.style.display = '';
    stopBtn.style.display  = '';
  } else if (status === 'paused') {
    startBtn.style.display = '';
    pauseBtn.style.display = 'none';
    stopBtn.style.display  = '';
    startBtn.textContent   = 'Resume';
  } else {
    startBtn.style.display = '';
    pauseBtn.style.display = 'none';
    stopBtn.style.display  = 'none';
    startBtn.textContent   = 'Start';
  }
}

// ── Toast system ─────────────────────────────────────────────

const toastQueue = [];
let toastSeq = 0;

function showToast(message, type = 'info') {
  // Synchronously evict oldest if at limit
  if (toastQueue.length >= 3) {
    const oldest = toastQueue.shift();
    clearTimeout(oldest.timer);
    const el = document.getElementById(`toast-${oldest.id}`);
    if (el) el.remove();
  }

  const id = ++toastSeq;
  const el = document.createElement('div');
  el.id        = `toast-${id}`;
  el.className = `toast toast-${type}`;
  el.textContent = message;
  el.addEventListener('click', () => dismissToast(id));
  toastContainer.appendChild(el);

  const timer = setTimeout(() => dismissToast(id), 4000);
  toastQueue.push({ id, timer });
}

function dismissToast(id) {
  const idx = toastQueue.findIndex(t => t.id === id);
  if (idx === -1) return;
  clearTimeout(toastQueue[idx].timer);
  toastQueue.splice(idx, 1);

  const el = document.getElementById(`toast-${id}`);
  if (!el) return;
  el.classList.add('is-leaving');
  el.addEventListener('animationend', () => el.remove(), { once: true });
}

// ── Native Windows toast ─────────────────────────────────────

function notify(title, message) {
  window.goNotify?.({ title, message })?.catch?.(() => {});
}

// ── Totals debounce ──────────────────────────────────────────

let totalsTimer = null;

function scheduleTotalsUpdate(totalElapsed) {
  if (totalsTimer) return;
  totalsTimer = setTimeout(async () => {
    totalsTimer = null;
    try {
      const { data } = await updateTotals({ totalElapsed, lastWaterAt, lastStretchAt });
      if (data) currentSession = data;
    } catch (err) {
      console.error('totals sync failed:', err);
    }
  }, 1000);
}

// ── Timer callbacks ──────────────────────────────────────────

timer.ontick = (remainingMs, progress) => {
  renderTick(remainingMs, progress);
  if (currentSession?.status !== 'running') return;

  const totalElapsed = currentSession.totalElapsed + timer.getCurrentElapsedMs() / 1000;

  if (currentSettings && totalElapsed - lastWaterAt >= currentSettings.waterInterval * 60) {
    showToast('Water \uD83D\uDCA7 \u2014 time to hydrate!', 'info');
    notify('\uD83D\uDCA7 Hydrate', 'Time to drink some water');
    lastWaterAt = totalElapsed;
    scheduleTotalsUpdate(totalElapsed);
  }
  if (currentSettings && totalElapsed - lastStretchAt >= currentSettings.stretchInterval * 60) {
    showToast('Stretch \uD83E\uDDD8 \u2014 time to move!', 'info');
    notify('\uD83E\uDDD8 Stretch', 'Time to move!');
    lastStretchAt = totalElapsed;
    scheduleTotalsUpdate(totalElapsed);
  }
};

timer.onsegmentend = async () => {
  if (!currentSession) { timer.reset(); return; }
  if (_segmentTransitioning) return;
  _segmentTransitioning = true;

  const oldIdx   = currentSession.segmentIndex;
  const newIdx   = (oldIdx + 1) % 8;
  const newType  = SEGMENT_SEQUENCE[newIdx];
  let   newCount = currentSession.pomodoroCount;

  if (SEGMENT_SEQUENCE[oldIdx] === 'focus') newCount += 1;
  if (newIdx === 0) newCount = 0;

  // Toast notification
  if (SEGMENT_SEQUENCE[oldIdx] === 'focus') {
    showToast(`Focus complete! ${newCount} / 4 pomodoros.`, 'success');
    notify('Focus complete! \uD83C\uDF45', `${newCount} / 4 pomodoros`);
  } else {
    showToast('Break over \u2014 back to work!', 'success');
    notify('Break over', 'Back to work!');
  }

  try {
    const { data } = await startSession({
      segmentType: newType, segmentIndex: newIdx, pomodoroCount: newCount,
    });
    if (data) {
      currentSession = data;
      lastWaterAt    = data.lastWaterAt;
      lastStretchAt  = data.lastStretchAt;
    }

    timer.newSegment(newType, currentSettings);
    timer.start();
    setSegmentLabel(newType);
    setPomodoroCount(newCount);
    triggerSegmentPulse();
    setControlState('running');
  } finally {
    _segmentTransitioning = false;
  }
};

// ── Control buttons ──────────────────────────────────────────

startBtn.addEventListener('click', async () => {
  const { data } = await startSession({
    segmentType:   currentSession?.segmentType  ?? 'focus',
    segmentIndex:  currentSession?.segmentIndex ?? 0,
    pomodoroCount: currentSession?.pomodoroCount ?? 0,
  });
  if (data) {
    currentSession = data;
    timer.start();
    setControlState('running');
  }
});

pauseBtn.addEventListener('click', async () => {
  timer.pause();
  const { data } = await pauseSession();
  if (data) currentSession = data;
  setControlState('paused');
});

stopBtn.addEventListener('click', async () => {
  timer.reset();
  const { data } = await stopSession();
  if (data) {
    currentSession = data;
    lastWaterAt    = data.lastWaterAt;
    lastStretchAt  = data.lastStretchAt;
  }
  setSegmentLabel('focus');
  setPomodoroCount(0);
  setControlState('idle');
  // Show full duration for focus segment
  const ms = (currentSettings?.pomodoroDuration ?? 25) * 60_000;
  renderTick(ms, 0);
});

settingsBtn.addEventListener('click', () => settings.open());
sortBtn.addEventListener('click', () => tasks.sortByPomodoros());
statsBtn?.addEventListener('click', () => statsPanel.open());

// ── Add task form ────────────────────────────────────────────

addTaskForm.addEventListener('submit', async e => {
  e.preventDefault();
  const titleEl    = addTaskForm.querySelector('#new-task-title');
  const pomEl      = addTaskForm.querySelector('#new-task-pomodoros');
  const priEl      = addTaskForm.querySelector('#new-task-priority');
  const catEl      = addTaskForm.querySelector('#new-task-category');
  const segEl      = addTaskForm.querySelector('#new-task-segment');

  const title = titleEl.value.trim();
  if (!title) return;

  const { data, error } = await tasks.addTask({
    title,
    estimatedPomodoros: Math.max(1, parseInt(pomEl.value, 10) || 1),
    priority: priEl.value,
    category: catEl.value.trim(),
    segmentMinutes: parseInt(segEl?.value, 10) || 0,
  });

  if (data) {
    addTaskForm.reset();
    priEl.value = 'medium';
    pomEl.value = '1';
    if (segEl) segEl.value = '';
    titleEl.focus();
    showToast('Task added', 'success');
  } else {
    showToast(`Could not add task: ${error}`, 'error');
  }
});

// ── Settings callback ────────────────────────────────────────

settings.onsave = s => {
  currentSettings = s;
  timer.updateDurations(s);
  tasks.updateSettings(s);
  showToast('Settings saved', 'success');
};

// ── Init ─────────────────────────────────────────────────────

async function init() {
  initTheme();
  initDesktop();
  initRuler();
  buildThemePicker(document.getElementById('theme-grid'));

  const [tasksRes, sessionRes, settingsRes] = await Promise.all([
    getTasks(), getSession(), getSettings(),
  ]);

  if (settingsRes.data) {
    currentSettings = settingsRes.data;
    settings.init(settingsRes.data);
  }

  if (tasksRes.data) tasks.init(tasksRes.data, currentSettings);

  if (sessionRes.data) {
    currentSession = sessionRes.data;
    lastWaterAt    = currentSession.lastWaterAt;
    lastStretchAt  = currentSession.lastStretchAt;

    setSegmentLabel(currentSession.segmentType);
    setPomodoroCount(currentSession.pomodoroCount);
    setControlState(currentSession.status);

    timer.init(currentSession, currentSettings ?? {
      pomodoroDuration: 25, shortBreak: 5, longBreak: 15,
    });
  }
}

// ── Visibility-change re-sync ─────────────────────────────────
// When the OS or browser hides the page (sleep, alt-tab, etc.) the RAF loop
// pauses. On return, we fetch a fresh session from the server and correct the
// timer's internal elapsed state so the countdown is immediately accurate.
// We skip re-syncs for very brief interruptions (< 2 s) to avoid jitter on
// quick alt-tabs.

let _hiddenAt = 0;

document.addEventListener('visibilitychange', async () => {
  if (document.visibilityState === 'hidden') {
    _hiddenAt = Date.now();
    return;
  }

  // Page became visible
  if (currentSession?.status !== 'running') return;
  if (_hiddenAt > 0 && Date.now() - _hiddenAt < 2000) return;

  try {
    const { data } = await getSession();
    if (!data) return;
    currentSession = data;
    lastWaterAt    = data.lastWaterAt;
    lastStretchAt  = data.lastStretchAt;
    timer.sync(data);
  } catch {
    // Silent fail — the timer will self-correct on the next user action
  }
});

document.addEventListener('DOMContentLoaded', init);
