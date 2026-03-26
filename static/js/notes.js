import { getNote, upsertNote } from './api.js';

const NOTE_DEBOUNCE_MS = 800;

export class NotesPanel {
  constructor() {
    this._textarea = null;
    this._debounce = null;
    this._date     = _todayKey();
  }

  /** Call once from app.js init() — wires the textarea auto-save */
  init() {
    this._textarea = document.getElementById('session-notes');
    if (!this._textarea) return;
    this._loadToday();
    this._textarea.addEventListener('input', () => this._scheduleUpsert());
  }

  /** Reload notes for today (call at midnight or session start) */
  async _loadToday() {
    this._date = _todayKey();
    const { data } = await getNote(this._date);
    if (this._textarea && data) this._textarea.value = data.text ?? '';
  }

  _scheduleUpsert() {
    clearTimeout(this._debounce);
    this._debounce = setTimeout(() => this._save(), NOTE_DEBOUNCE_MS);
  }

  async _save() {
    const text = this._textarea?.value ?? '';
    await upsertNote(this._date, text);
  }
}

function _todayKey() {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`;
}
