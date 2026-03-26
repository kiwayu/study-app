'use client'

import { useCallback, useEffect, useRef, useState } from 'react'
import type { Session, Settings } from '@/lib/supabase/types'

export const SEGMENT_SEQUENCE = [
  'focus', 'short_break', 'focus', 'short_break',
  'focus', 'short_break', 'focus', 'long_break',
] as const

function durationMs(type: string, settings: Settings): number {
  switch (type) {
    case 'focus':       return settings.pomodoro_duration * 60_000
    case 'short_break': return settings.short_break * 60_000
    case 'long_break':  return settings.long_break * 60_000
    default:            return settings.pomodoro_duration * 60_000
  }
}

function fmtTime(ms: number): string {
  const s = Math.max(0, Math.ceil(ms / 1000))
  return `${String(Math.floor(s / 60)).padStart(2, '0')}:${String(s % 60).padStart(2, '0')}`
}

interface TimerState {
  remainingMs: number
  progress: number
  formattedTime: string
}

interface UseTimerReturn {
  timerState: TimerState
  start: () => void
  pause: () => void
  reset: () => void
  sync: (session: Session) => void
  getCurrentElapsedMs: () => number
}

export function useTimer(
  initialSession: Session,
  settings: Settings,
  onSegmentEnd: (newIndex: number, newType: string, newCount: number) => void
): UseTimerReturn {
  const segmentDurationMsRef = useRef(durationMs(initialSession.segment_type, settings))
  const bankedMsRef          = useRef(0)
  const rafIdRef             = useRef<number | null>(null)
  const rafStartTimeRef      = useRef(0)
  const transitioning        = useRef(false)

  const [timerState, setTimerState] = useState<TimerState>(() => {
    const dur = segmentDurationMsRef.current
    return { remainingMs: dur, progress: 0, formattedTime: fmtTime(dur) }
  })

  const emitTick = useCallback((elapsedMs: number) => {
    const dur       = segmentDurationMsRef.current
    const clamped   = Math.min(elapsedMs, dur)
    const remaining = dur - clamped
    const progress  = dur > 0 ? clamped / dur : 0
    setTimerState({ remainingMs: remaining, progress, formattedTime: fmtTime(remaining) })
  }, [])

  // Refs for mutable session state accessed inside RAF callback
  const segIndexRef = useRef(initialSession.segment_index)
  const pomCountRef = useRef(initialSession.pomodoro_count)
  const onSegmentEndRef = useRef(onSegmentEnd)
  useEffect(() => { onSegmentEndRef.current = onSegmentEnd })

  const startRAF = useCallback(() => {
    rafStartTimeRef.current = performance.now()
    const tick = () => {
      const elapsed   = bankedMsRef.current + (performance.now() - rafStartTimeRef.current)
      const remaining = segmentDurationMsRef.current - elapsed

      if (remaining <= 0) {
        bankedMsRef.current = segmentDurationMsRef.current
        rafIdRef.current    = null
        setTimerState({ remainingMs: 0, progress: 1, formattedTime: '00:00' })
        if (!transitioning.current) {
          transitioning.current = true
          const newIdx   = (segIndexRef.current + 1) % 8
          const newType  = SEGMENT_SEQUENCE[newIdx]
          let   newCount = pomCountRef.current
          if (SEGMENT_SEQUENCE[segIndexRef.current] === 'focus') newCount++
          if (newIdx === 0) newCount = 0
          onSegmentEndRef.current(newIdx, newType, newCount)
          transitioning.current = false
        }
        return
      }

      const progress = elapsed / segmentDurationMsRef.current
      setTimerState({ remainingMs: remaining, progress, formattedTime: fmtTime(remaining) })
      rafIdRef.current = requestAnimationFrame(tick)
    }
    rafIdRef.current = requestAnimationFrame(tick)
  }, [])

  // Initialise from server session state
  useEffect(() => {
    const { status, started_at, banked_ms, segment_type, segment_index, pomodoro_count } = initialSession
    segmentDurationMsRef.current = durationMs(segment_type, settings)
    segIndexRef.current          = segment_index
    pomCountRef.current          = pomodoro_count

    if (status === 'running' && started_at) {
      bankedMsRef.current = banked_ms + (Date.now() - new Date(started_at).getTime())
    } else if (status === 'paused') {
      bankedMsRef.current = banked_ms
    } else {
      bankedMsRef.current = 0
    }
    bankedMsRef.current = Math.max(0, bankedMsRef.current)

    emitTick(bankedMsRef.current)

    if (status === 'running') {
      if (bankedMsRef.current >= segmentDurationMsRef.current) {
        requestAnimationFrame(() => {
          const newIdx   = (segment_index + 1) % 8
          const newType  = SEGMENT_SEQUENCE[newIdx]
          let newCount   = pomodoro_count
          if (SEGMENT_SEQUENCE[segment_index] === 'focus') newCount++
          if (newIdx === 0) newCount = 0
          onSegmentEndRef.current(newIdx, newType, newCount)
        })
      } else {
        startRAF()
      }
    }

    return () => {
      if (rafIdRef.current) cancelAnimationFrame(rafIdRef.current)
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []) // Only on mount — subsequent updates go through start/pause/sync

  const start = useCallback(() => {
    if (rafIdRef.current) return
    startRAF()
  }, [startRAF])

  const pause = useCallback(() => {
    if (!rafIdRef.current) return
    bankedMsRef.current += performance.now() - rafStartTimeRef.current
    cancelAnimationFrame(rafIdRef.current)
    rafIdRef.current = null
  }, [])

  const reset = useCallback(() => {
    if (rafIdRef.current) { cancelAnimationFrame(rafIdRef.current); rafIdRef.current = null }
    bankedMsRef.current = 0
    emitTick(0)
  }, [emitTick])

  const sync = useCallback((session: Session) => {
    if (session.status !== 'running') return
    const serverElapsed = session.banked_ms + (Date.now() - new Date(session.started_at!).getTime())
    bankedMsRef.current          = Math.max(0, serverElapsed)
    rafStartTimeRef.current      = performance.now()
    segIndexRef.current          = session.segment_index
    pomCountRef.current          = session.pomodoro_count
    segmentDurationMsRef.current = durationMs(session.segment_type, settings)
  }, [settings])

  const getCurrentElapsedMs = useCallback(() => {
    if (rafIdRef.current) {
      return bankedMsRef.current + (performance.now() - rafStartTimeRef.current)
    }
    return bankedMsRef.current
  }, [])

  return { timerState, start, pause, reset, sync, getCurrentElapsedMs }
}
