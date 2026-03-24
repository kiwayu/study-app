const SEGMENT_SEQUENCE = [
  'focus', 'short_break', 'focus', 'short_break',
  'focus', 'short_break', 'focus', 'long_break',
];

export { SEGMENT_SEQUENCE };

export class TimerEngine {
  constructor() {
    this._segmentDurationMs = 25 * 60 * 1000;
    this._segmentType       = 'focus';
    this._bankedMs          = 0;
    this._rafId             = null;
    this._rafStartTime      = 0;

    /** @type {(remainingMs: number, progress: number) => void} */
    this.ontick = null;
    /** @type {() => void} */
    this.onsegmentend = null;
  }

  _durationMs(type, settings) {
    switch (type) {
      case 'focus':       return settings.pomodoroDuration * 60_000;
      case 'short_break': return settings.shortBreak        * 60_000;
      case 'long_break':  return settings.longBreak         * 60_000;
      default:            return settings.pomodoroDuration * 60_000;
    }
  }

  /**
   * Initialise from server-side SessionState.
   * Always performs one synchronous render pass before starting the RAF loop.
   */
  init(sessionState, settings) {
    this._segmentType       = sessionState.segmentType ?? 'focus';
    this._segmentDurationMs = this._durationMs(this._segmentType, settings);

    const { status, elapsedSeconds, startedAt } = sessionState;

    if (status === 'running' && startedAt) {
      this._bankedMs = elapsedSeconds * 1000
        + (Date.now() - new Date(startedAt).getTime());
    } else if (status === 'paused') {
      this._bankedMs = elapsedSeconds * 1000;
    } else {
      this._bankedMs = 0;
    }
    this._bankedMs = Math.max(0, this._bankedMs);

    // Synchronous initial render
    this._emitTick(this._bankedMs);

    if (status === 'running') {
      if (this._bankedMs >= this._segmentDurationMs) {
        // Segment expired while page was away — fire after 1 frame
        requestAnimationFrame(() => { if (this.onsegmentend) this.onsegmentend(); });
      } else {
        this._startRAF();
      }
    }
  }

  /**
   * Prepare for a new segment. Resets banked time and re-renders display.
   * NOTE: This method extends the spec's public API (which listed `reset` + `updateDurations`).
   * It is a necessary addition: `reset()` alone does not update `_segmentType`, so the duration
   * for the next segment would be calculated against the wrong type. `newSegment()` atomically
   * sets type + duration + resets elapsed in one call, which is what `onsegmentend` in app.js needs.
   */
  newSegment(type, settings) {
    if (this._rafId) { cancelAnimationFrame(this._rafId); this._rafId = null; }
    this._segmentType       = type;
    this._segmentDurationMs = this._durationMs(type, settings);
    this._bankedMs          = 0;
    this._emitTick(0);
  }

  /** Start the RAF loop from current banked position. */
  start() {
    if (this._rafId) return;
    this._startRAF();
  }

  /** Stop the RAF loop and bank current elapsed. */
  pause() {
    if (!this._rafId) return;
    this._bankedMs += performance.now() - this._rafStartTime;
    cancelAnimationFrame(this._rafId);
    this._rafId = null;
  }

  /** Stop and reset to segment start. */
  reset() {
    if (this._rafId) { cancelAnimationFrame(this._rafId); this._rafId = null; }
    this._bankedMs = 0;
    this._emitTick(0);
  }

  /** Recalculate duration for the current segment type without resetting elapsed. */
  updateDurations(settings) {
    this._segmentDurationMs = this._durationMs(this._segmentType, settings);
  }

  /** Current elapsed ms in this segment (including running time if active). */
  getCurrentElapsedMs() {
    if (this._rafId) {
      return this._bankedMs + (performance.now() - this._rafStartTime);
    }
    return this._bankedMs;
  }

  // ── Private ──────────────────────────────────────────────

  _startRAF() {
    this._rafStartTime = performance.now();
    const tick = () => {
      const elapsed   = this._bankedMs + (performance.now() - this._rafStartTime);
      const remaining = this._segmentDurationMs - elapsed;

      if (remaining <= 0) {
        this._bankedMs = this._segmentDurationMs;
        this._rafId    = null;
        if (this.ontick)       this.ontick(0, 1);
        if (this.onsegmentend) this.onsegmentend();
        return;
      }

      const progress = elapsed / this._segmentDurationMs;
      if (this.ontick) this.ontick(remaining, progress);
      this._rafId = requestAnimationFrame(tick);
    };
    this._rafId = requestAnimationFrame(tick);
  }

  _emitTick(elapsedMs) {
    const clamped   = Math.min(elapsedMs, this._segmentDurationMs);
    const remaining = this._segmentDurationMs - clamped;
    const progress  = this._segmentDurationMs > 0 ? clamped / this._segmentDurationMs : 0;
    if (this.ontick) this.ontick(remaining, progress);
  }
}
