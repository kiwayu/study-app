'use client'

import { useCallback, useState } from 'react'
import { createClient } from '@/lib/supabase/client'
import { useRealtime } from './useRealtime'
import type { Session } from '@/lib/supabase/types'

export function useSession(initialSession: Session, userId: string) {
  const [session, setSession] = useState<Session>(initialSession)
  const supabase = createClient()

  // Real-time subscription — another device changed session state
  useRealtime<Record<string, unknown>>({
    table: 'session',
    filter: `user_id=eq.${userId}`,
    onUpdate: (row) => setSession(row as unknown as Session),
  })

  const startSession = useCallback(async (updates: Partial<Session>) => {
    const now = new Date().toISOString()
    const changes = { ...updates, status: 'running' as const, started_at: now }
    const prev = session
    setSession(s => ({ ...s, ...changes }))

    const { error } = await supabase.from('session').update(changes).eq('user_id', userId)
    if (error) setSession(prev) // rollback
  }, [supabase, session, userId])

  const pauseSession = useCallback(async (bankedMs: number) => {
    const changes = { status: 'paused' as const, banked_ms: bankedMs, started_at: null }
    const prev = session
    setSession(s => ({ ...s, ...changes }))

    const { error } = await supabase.from('session').update(changes).eq('user_id', userId)
    if (error) setSession(prev)
  }, [supabase, session, userId])

  const stopSession = useCallback(async () => {
    const changes = { status: 'idle' as const, banked_ms: 0, started_at: null, segment_index: 0, segment_type: 'focus' as const, pomodoro_count: 0 }
    const prev = session
    setSession(s => ({ ...s, ...changes }))

    const { error } = await supabase.from('session').update(changes).eq('user_id', userId)
    if (error) setSession(prev)
  }, [supabase, session, userId])

  const updateTotals = useCallback(async (totals: Session['totals']) => {
    await supabase.from('session').update({ totals }).eq('user_id', userId)
  }, [supabase, userId])

  return { session, startSession, pauseSession, stopSession, updateTotals, setSession }
}
