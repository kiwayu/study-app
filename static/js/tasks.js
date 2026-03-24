import { createTask, updateTask, deleteTask } from './api.js';

const PRIORITY_CYCLE = ['high', 'medium', 'low'];

export class TaskManager {
  constructor(containerEl) {
    this._el    = containerEl;
    this._tasks = [];
    this._dragSrcIdx    = null;
    this._dropOccurred  = false;

    /** @type {(tasks: object[]) => void} */
    this.onchange = null;
  }

  init(tasks) {
    this._tasks = [...tasks].sort((a, b) => a.order - b.order);
    this._render();
  }

  async addTask(req) {
    const { data, error } = await createTask(req);
    if (data) {
      this._tasks.push(data);
      this._render();
      if (this.onchange) this.onchange([...this._tasks]);
    }
    return { data, error };
  }

  async sortByPomodoros() {
    const pri = { high: 0, medium: 1, low: 2 };
    this._tasks.sort((a, b) =>
      a.estimatedPomodoros !== b.estimatedPomodoros
        ? a.estimatedPomodoros - b.estimatedPomodoros
        : pri[a.priority] - pri[b.priority]
    );
    await this._reindex();
    this._render();
  }

  // ── Render ─────────────────────────────────────────────────

  _render() {
    this._el.innerHTML = '';
    this._tasks.forEach((task, idx) => {
      this._el.appendChild(this._buildCard(task, idx));
    });
  }

  _buildCard(task, idx) {
    const el = document.createElement('div');
    el.className = 'task-card' + (task.completed ? ' is-completed' : '');
    el.dataset.id  = task.id;
    el.dataset.idx = String(idx);
    el.setAttribute('role', 'listitem');
    el.draggable = true;

    el.innerHTML = `
      <div class="task-drag-handle" aria-hidden="true">⠿</div>
      <div class="task-main">
        <div class="task-top">
          <input type="checkbox" class="task-checkbox" ${task.completed ? 'checked' : ''}
                 aria-label="Mark complete">
          <span class="task-title">${esc(task.title)}</span>
        </div>
        <div class="task-meta">
          <span class="task-priority priority-${task.priority}">${task.priority}</span>
          <span class="task-pomodoros">×${task.estimatedPomodoros}</span>
          ${task.category ? `<span class="task-category">${esc(task.category)}</span>` : ''}
        </div>
      </div>
      <button class="task-delete" aria-label="Delete task" tabindex="-1">×</button>
    `;

    // Events
    el.querySelector('.task-checkbox').addEventListener('change',
      () => this._toggleComplete(task.id));
    el.querySelector('.task-title').addEventListener('click',
      e => this._editTitle(e.currentTarget, task.id));
    el.querySelector('.task-priority').addEventListener('click',
      e => this._cyclePriority(e.currentTarget, task.id));
    el.querySelector('.task-pomodoros').addEventListener('click',
      e => this._editPomodoros(e.currentTarget, task.id));
    el.querySelector('.task-delete').addEventListener('click',
      () => this._deleteTask(task.id, el));

    // Desktop DnD
    el.addEventListener('dragstart', e => this._onDragStart(e, idx));
    el.addEventListener('dragover',  e => this._onDragOver(e));
    el.addEventListener('drop',      e => this._onDrop(e, idx));
    el.addEventListener('dragend',   e => this._onDragEnd(e));

    // Mobile touch DnD
    el.addEventListener('touchstart', e => this._onTouchStart(e, idx), { passive: false });

    return el;
  }

  // ── Interactions ───────────────────────────────────────────

  async _toggleComplete(id) {
    const task = this._find(id);
    if (!task) return;
    const { data } = await updateTask(id, { completed: !task.completed });
    if (data) {
      this._replace(id, data);
      this._render();
      if (this.onchange) this.onchange([...this._tasks]);
    }
  }

  _editTitle(span, id) {
    const orig = span.textContent;
    const input = Object.assign(document.createElement('input'), {
      className: 'task-title-edit',
      value: orig,
    });
    span.replaceWith(input);
    input.focus();
    input.select();

    let done = false;
    const save = async () => {
      if (done) return; done = true;
      const val = input.value.trim() || orig;
      const { data } = await updateTask(id, { title: val });
      const newSpan = makeSpan('task-title', data?.title ?? val);
      newSpan.addEventListener('click', e => this._editTitle(e.currentTarget, id));
      input.replaceWith(newSpan);
      if (data) this._replace(id, data);
    };
    const cancel = () => {
      if (done) return; done = true;
      const newSpan = makeSpan('task-title', orig);
      newSpan.addEventListener('click', e => this._editTitle(e.currentTarget, id));
      input.replaceWith(newSpan);
    };

    input.addEventListener('blur', save);
    input.addEventListener('keydown', e => {
      if (e.key === 'Enter')  { e.preventDefault(); save(); }
      if (e.key === 'Escape') { e.preventDefault(); cancel(); }
    });
  }

  async _cyclePriority(badge, id) {
    const task = this._find(id);
    if (!task) return;
    const next = PRIORITY_CYCLE[(PRIORITY_CYCLE.indexOf(task.priority) + 1) % 3];
    const { data } = await updateTask(id, { priority: next });
    if (data) {
      this._replace(id, data);
      badge.textContent = next;
      badge.className   = `task-priority priority-${next}`;
    }
  }

  _editPomodoros(badge, id) {
    const task = this._find(id);
    if (!task) return;
    const orig = task.estimatedPomodoros;
    const input = Object.assign(document.createElement('input'), {
      type: 'number', min: '1', max: '9',
      className: 'task-pomodoros-edit',
      value: String(orig),
    });
    badge.replaceWith(input);
    input.focus();
    input.select();

    let done = false;
    const save = async () => {
      if (done) return; done = true;
      const val = Math.max(1, Math.min(9, parseInt(input.value, 10) || orig));
      const { data } = await updateTask(id, { estimatedPomodoros: val });
      const newBadge = makeSpan('task-pomodoros', `×${data?.estimatedPomodoros ?? val}`);
      newBadge.addEventListener('click', e => this._editPomodoros(e.currentTarget, id));
      input.replaceWith(newBadge);
      if (data) this._replace(id, data);
    };
    const cancel = () => {
      if (done) return; done = true;
      const newBadge = makeSpan('task-pomodoros', `×${orig}`);
      newBadge.addEventListener('click', e => this._editPomodoros(e.currentTarget, id));
      input.replaceWith(newBadge);
    };

    input.addEventListener('blur', save);
    input.addEventListener('keydown', e => {
      if (e.key === 'Enter')  { e.preventDefault(); save(); }
      if (e.key === 'Escape') { e.preventDefault(); cancel(); }
    });
  }

  _deleteTask(id, el) {
    el.classList.add('is-deleting');
    el.addEventListener('animationend', async () => {
      const { error } = await deleteTask(id);
      if (!error) {
        this._tasks = this._tasks.filter(t => t.id !== id);
        el.remove();
        if (this.onchange) this.onchange([...this._tasks]);
      }
    }, { once: true });
  }

  // ── Desktop DnD ────────────────────────────────────────────

  _onDragStart(e, idx) {
    this._dragSrcIdx   = idx;
    this._dropOccurred = false;
    e.dataTransfer.effectAllowed = 'move';
    e.currentTarget.classList.add('is-dragging');
  }

  _onDragOver(e) {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    // Highlight only this target
    this._el.querySelectorAll('.task-card').forEach(c => c.classList.remove('is-drop-target'));
    e.currentTarget.classList.add('is-drop-target');
  }

  async _onDrop(e, targetIdx) {
    e.preventDefault();
    this._dropOccurred = true;
    this._el.querySelectorAll('.task-card').forEach(c => c.classList.remove('is-drop-target'));
    if (this._dragSrcIdx === targetIdx) return;

    const [moved] = this._tasks.splice(this._dragSrcIdx, 1);
    this._tasks.splice(targetIdx, 0, moved);
    await this._reindex();
    this._render();
    if (this.onchange) this.onchange([...this._tasks]);
  }

  _onDragEnd(e) {
    e.currentTarget.classList.remove('is-dragging');
    this._el.querySelectorAll('.task-card').forEach(c => c.classList.remove('is-drop-target'));
    if (!this._dropOccurred) this._render(); // cancelled — restore DOM
    this._dragSrcIdx = null;
  }

  // ── Mobile touch DnD ───────────────────────────────────────

  _onTouchStart(e, srcIdx) {
    if (e.touches.length !== 1) return;
    e.preventDefault();

    const srcEl  = e.currentTarget;
    const rect   = srcEl.getBoundingClientRect();
    const touch  = e.touches[0];
    const offX   = touch.clientX - rect.left;
    const offY   = touch.clientY - rect.top;

    const ghost = srcEl.cloneNode(true);
    Object.assign(ghost.style, {
      position: 'fixed', width: `${rect.width}px`, pointerEvents: 'none',
      zIndex: '999', top: `${rect.top}px`, left: `${rect.left}px`,
      opacity: '0.85', boxShadow: 'var(--shadow-drag)',
    });
    document.body.appendChild(ghost);

    let dropIdx = srcIdx;

    const onMove = mv => {
      const t = mv.touches[0];
      ghost.style.top  = `${t.clientY - offY}px`;
      ghost.style.left = `${t.clientX - offX}px`;

      ghost.style.display = 'none';
      const below = document.elementFromPoint(t.clientX, t.clientY);
      ghost.style.display = '';

      const card = below?.closest('.task-card');
      this._el.querySelectorAll('.task-card').forEach(c => c.classList.remove('is-drop-target'));
      if (card && card !== srcEl) {
        card.classList.add('is-drop-target');
        dropIdx = parseInt(card.dataset.idx, 10);
      }
    };

    const finish = async () => {
      cleanup();
      ghost.remove();
      this._el.querySelectorAll('.task-card').forEach(c => c.classList.remove('is-drop-target'));
      if (dropIdx !== srcIdx) {
        const [moved] = this._tasks.splice(srcIdx, 1);
        this._tasks.splice(dropIdx, 0, moved);
        await this._reindex();
        this._render();
        if (this.onchange) this.onchange([...this._tasks]);
      }
    };

    const cancel = () => {
      cleanup();
      ghost.remove();
      this._el.querySelectorAll('.task-card').forEach(c => c.classList.remove('is-drop-target'));
    };

    const cleanup = () => {
      document.removeEventListener('touchmove',   onMove,  { passive: false });
      document.removeEventListener('touchend',    finish);
      document.removeEventListener('touchcancel', cancel);
    };

    document.addEventListener('touchmove',   onMove,  { passive: false });
    document.addEventListener('touchend',    finish);
    document.addEventListener('touchcancel', cancel);
  }

  // ── Helpers ────────────────────────────────────────────────

  _find(id)          { return this._tasks.find(t => t.id === id) ?? null; }
  _replace(id, data) {
    const i = this._tasks.findIndex(t => t.id === id);
    if (i !== -1) this._tasks[i] = data;
  }

  async _reindex() {
    const updates = this._tasks.map((t, i) => {
      if (t.order !== i) { t.order = i; return updateTask(t.id, { order: i }); }
      return null;
    }).filter(Boolean);
    await Promise.all(updates);
  }
}

// ── Utilities ───────────────────────────────────────────────

function esc(s) {
  return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

function makeSpan(cls, text) {
  const s = document.createElement('span');
  s.className   = cls;
  s.textContent = text;
  return s;
}
