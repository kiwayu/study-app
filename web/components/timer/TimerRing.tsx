interface TimerRingProps {
  progress: number  // 0–1
  formattedTime: string
  segmentType: string
  pomodoroCount: number
  status: 'idle' | 'running' | 'paused'
}

const CIRCUMFERENCE = 339.292  // 2π × 54

const SEGMENT_LABELS: Record<string, string> = {
  focus: 'Focus',
  short_break: 'Short Break',
  long_break: 'Long Break',
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function TimerRing({ progress, formattedTime, segmentType, pomodoroCount, status }: TimerRingProps) {
  const offset = CIRCUMFERENCE * (1 - progress)

  return (
    <div className="flex flex-col items-center gap-3 select-none">
      {/* SVG Ring */}
      <div className="relative w-44 h-44">
        <svg viewBox="0 0 120 120" className="w-full h-full -rotate-90">
          {/* Track */}
          <circle
            cx="60" cy="60" r="54"
            fill="none"
            stroke="var(--color-surface-2)"
            strokeWidth="6"
          />
          {/* Progress */}
          <circle
            cx="60" cy="60" r="54"
            fill="none"
            stroke="var(--color-accent)"
            strokeWidth="6"
            strokeLinecap="round"
            strokeDasharray={CIRCUMFERENCE}
            strokeDashoffset={offset}
            style={{ transition: 'stroke-dashoffset 0.3s ease' }}
          />
        </svg>

        {/* Time display inside ring */}
        <div className="absolute inset-0 flex flex-col items-center justify-center gap-0.5">
          <span className="text-3xl font-semibold tracking-tight tabular-nums text-zinc-50">
            {formattedTime}
          </span>
          <span className="text-xs font-medium text-[var(--color-muted)] tracking-wide uppercase">
            {SEGMENT_LABELS[segmentType] ?? segmentType}
          </span>
        </div>
      </div>

      {/* Pomodoro count */}
      <div className="text-xs text-[var(--color-muted)] font-medium tracking-widest">
        {pomodoroCount} / 4
      </div>
    </div>
  )
}
