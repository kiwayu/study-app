import { updateSettings } from './api.js';

export class SettingsManager {
  constructor(drawerEl, overlayEl) {
    this._drawer  = drawerEl;
    this._overlay = overlayEl;
    this._current = null;

    /** @type {(settings: object) => void} */
    this.onsave = null;

    // All [data-cancel] buttons close the drawer
    this._drawer.querySelectorAll('[data-cancel]').forEach(btn => {
      btn.addEventListener('click', () => this.close());
    });
    this._overlay.addEventListener('click', () => this.close());
    this._drawer.querySelector('#settings-save').addEventListener('click', () => this._save());

    // Keyboard: Escape closes
    document.addEventListener('keydown', e => {
      if (e.key === 'Escape' && this._drawer.classList.contains('is-open')) this.close();
    });
  }

  init(settings) {
    this._current = { ...settings };
    this._populate(settings);
  }

  open() {
    this._populate(this._current);
    this._drawer.classList.add('is-open');
    this._overlay.classList.add('is-open');
    // Focus first input for accessibility
    const first = this._drawer.querySelector('input');
    if (first) setTimeout(() => first.focus(), 320);
  }

  close() {
    this._drawer.classList.remove('is-open');
    this._overlay.classList.remove('is-open');
  }

  // ── Private ─────────────────────────────────────────────────

  _populate(s) {
    this._val('#s-pomodoro', s.pomodoroDuration);
    this._val('#s-short',    s.shortBreak);
    this._val('#s-long',     s.longBreak);
    this._val('#s-water',    s.waterInterval);
    this._val('#s-stretch',  s.stretchInterval);
  }

  _val(selector, value) {
    const el = this._drawer.querySelector(selector);
    if (el) el.value = String(value);
  }

  _num(selector, fallback) {
    const el = this._drawer.querySelector(selector);
    return Math.max(1, parseInt(el?.value ?? '', 10) || fallback);
  }

  async _save() {
    const settings = {
      pomodoroDuration: this._num('#s-pomodoro', 25),
      shortBreak:       this._num('#s-short',     5),
      longBreak:        this._num('#s-long',      15),
      waterInterval:    this._num('#s-water',     45),
      stretchInterval:  this._num('#s-stretch',   60),
    };

    const { data, error } = await updateSettings(settings);
    if (data) {
      this._current = data;
      this.close();
      if (this.onsave) this.onsave(data);
    } else {
      console.error('settings save failed:', error);
    }
  }
}
