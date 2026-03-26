'use client'

import { useEffect } from 'react'
import { createClient } from '@/lib/supabase/client'
import type { RealtimePostgresChangesPayload } from '@supabase/supabase-js'

type ChangeEvent = 'INSERT' | 'UPDATE' | 'DELETE' | '*'

interface UseRealtimeOptions<T extends Record<string, unknown>> {
  table: string
  filter?: string
  event?: ChangeEvent
  onInsert?: (row: T) => void
  onUpdate?: (row: T) => void
  onDelete?: (row: T) => void
}

export function useRealtime<T extends Record<string, unknown>>({
  table,
  filter,
  event = '*',
  onInsert,
  onUpdate,
  onDelete,
}: UseRealtimeOptions<T>) {
  useEffect(() => {
    const supabase = createClient()
    const channel = supabase
      .channel(`realtime:${table}`)
      .on(
        'postgres_changes',
        {
          event,
          schema: 'public',
          table,
          ...(filter ? { filter } : {}),
        },
        (payload: RealtimePostgresChangesPayload<T>) => {
          if (payload.eventType === 'INSERT' && onInsert) {
            onInsert(payload.new as T)
          } else if (payload.eventType === 'UPDATE' && onUpdate) {
            onUpdate(payload.new as T)
          } else if (payload.eventType === 'DELETE' && onDelete) {
            onDelete(payload.old as T)
          }
        }
      )
      .subscribe()

    // Cleanup: unsubscribe on unmount to prevent memory leaks and duplicate subscriptions
    return () => {
      supabase.removeChannel(channel)
    }
  }, [table, filter, event, onInsert, onUpdate, onDelete])
}
