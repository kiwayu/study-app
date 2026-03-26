'use client'

import { useEffect, useRef } from 'react'
import { createClient } from '@/lib/supabase/client'
import type { RealtimePostgresChangesPayload } from '@supabase/supabase-js'

type ChangeEvent = 'INSERT' | 'UPDATE' | 'DELETE' | '*'

interface UseRealtimeOptions<T extends Record<string, unknown>> {
  table: string
  filter?: string
  event?: ChangeEvent
  onInsert?: (row: T) => void
  onUpdate?: (row: T) => void
  onDelete?: (row: Partial<T>) => void
}

export function useRealtime<T extends Record<string, unknown>>({
  table,
  filter,
  event = '*',
  onInsert,
  onUpdate,
  onDelete,
}: UseRealtimeOptions<T>) {
  // Stable refs for callbacks — updated every render but never trigger re-subscription
  const onInsertRef = useRef(onInsert)
  const onUpdateRef = useRef(onUpdate)
  const onDeleteRef = useRef(onDelete)
  useEffect(() => { onInsertRef.current = onInsert })
  useEffect(() => { onUpdateRef.current = onUpdate })
  useEffect(() => { onDeleteRef.current = onDelete })

  // Stable unique channel name — generated once on mount
  const channelNameRef = useRef(
    `realtime:${table}:${event}:${filter ?? 'all'}:${Math.random().toString(36).slice(2)}`
  )

  useEffect(() => {
    const supabase = createClient()
    const channel = supabase
      .channel(channelNameRef.current)
      .on(
        'postgres_changes',
        {
          event,
          schema: 'public',
          table,
          ...(filter ? { filter } : {}),
        },
        (payload: RealtimePostgresChangesPayload<T>) => {
          if (payload.eventType === 'INSERT' && onInsertRef.current) {
            onInsertRef.current(payload.new as T)
          } else if (payload.eventType === 'UPDATE' && onUpdateRef.current) {
            onUpdateRef.current(payload.new as T)
          } else if (payload.eventType === 'DELETE' && onDeleteRef.current) {
            onDeleteRef.current(payload.old as Partial<T>)
          }
        }
      )
      .subscribe()

    return () => {
      supabase.removeChannel(channel)
    }
  // table, filter, and event are stable identifiers — changing them creates a new subscription intentionally
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [table, filter, event])
}
