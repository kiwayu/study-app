'use client'

import { useCallback, useEffect, useRef } from 'react'
import { TimerRing } from './TimerRing'
import { useTimer } from '@/lib/hooks/useTimer'
import type { Session, Settings } from '@/lib/supabase/types'

interface TimerPanelProps {
  session: Session
  settings: Settings
  onStart: (updates: Partial<Session>) => void
  onPause: (bankedMs: number) => void
  onStop: () => void
  onSettingsOpen: () => void
}

export function TimerPanel({
  session,
  settings,
  onStart,
  onPause,
  onStop,
  onSettingsOpen,
}: TimerPanelProps) {
  // Stable initial session for useTimer (only used on mount)
  const initialSessionRef = useRef(session)

  const handleSegmentEnd = useCallback((newIndex: number, newType: string, newCount: number) => {
    onStart({
      segment_index: newIndex,
      segment_type: newType as Session['segment_type'],
      pomodoro_count: newCount,
    })
  }, [onStart])

  const { timerState, start, pause, reset, getCurrentElapsedMs } = useTimer(
    initialSessionRef.current,
    settings,
    handleSegmentEnd
  )

  // Live values from current session — drives button rendering and keyboard logic
  const { status, segment_type, segment_index, pomodoro_count } = session

  const handleStart = useCallback(() => {
    start()
    onStart({ segment_type, segment_index, pomodoro_count })
  }, [start, onStart, segment_type, segment_index, pomodoro_count])

  const handlePause = useCallback(() => {
    pause()
    onPause(getCurrentElapsedMs())
  }, [pause, onPause, getCurrentElapsedMs])

  const handleStop = useCallback(() => {
    reset()
    onStop()
  }, [reset, onStop])

  // Listen for keyboard Space shortcut dispatched from AppShell
  useEffect(() => {
    function onSpaceStart() { if (status !== 'running') handleStart() }
    function onSpacePause() { if (status === 'running') handlePause() }
    document.addEventListener('timer:start', onSpaceStart)
    document.addEventListener('timer:pause', onSpacePause)
    return () => {
      document.removeEventListener('timer:start', onSpaceStart)
      document.removeEventListener('timer:pause', onSpacePause)
    }
  }, [status, handleStart, handlePause])

  return (
    <aside
      className="flex flex-col items-center gap-6 py-8 px-5 shrink-0"
      style={{ width: `${settings.sidebar_width}px` }}
    >
      {/* Timer ring */}
      <TimerRing
        progress={timerState.progress}
        formattedTime={timerState.formattedTime}
        segmentType={segment_type}
        pomodoroCount={pomodoro_count}
        status={status}
      />

      {/* Controls */}
      <div className="flex items-center gap-2 w-full">
        {status !== 'running' ? (
          <button
            onClick={handleStart}
            className="flex-1 rounded-full bg-[var(--color-accent)] hover:bg-[var(--color-accent-dim)] text-white text-sm font-semibold py-2.5 transition-colors duration-150"
          >
            {status === 'paused' ? 'Resume' : 'Start'}
          </button>
        ) : (
          <button
            onClick={handlePause}
            className="flex-1 rounded-full bg-[var(--color-surface-2)] hover:bg-zinc-600 text-zinc-200 text-sm font-semibold py-2.5 transition-colors duration-150"
          >
            Pause
          </button>
        )}

        {status !== 'idle' && (
          <button
            onClick={handleStop}
            className="rounded-full bg-[var(--color-surface-2)] hover:bg-zinc-600 text-zinc-400 text-sm font-semibold px-4 py-2.5 transition-colors duration-150"
          >
            Stop
          </button>
        )}
      </div>

      {/* Icon buttons row */}
      <div className="flex items-center justify-center gap-4 w-full mt-auto">
        <button
          onClick={onSettingsOpen}
          aria-label="Open settings"
          className="p-2 rounded-lg text-[var(--color-muted)] hover:text-zinc-200 hover:bg-[var(--color-surface-2)] transition-colors"
        >
          <svg viewBox="0 0 20 20" fill="currentColor" className="w-5 h-5" aria-hidden="true">
            <path fillRule="evenodd" d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z" clipRule="evenodd" />
          </svg>
        </button>
      </div>
    </aside>
  )
}
