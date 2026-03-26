'use client'

import { useCallback, useMemo, useState } from 'react'
import { createClient } from '@/lib/supabase/client'
import type { Completion } from '@/lib/supabase/types'

export function useStats() {
  const [completions, setCompletions] = useState<Completion[]>([])
  const [loading, setLoading] = useState(false)
  const supabase = useMemo(() => createClient(), [])

  const fetchCompletions = useCallback(async (days = 365) => {
    setLoading(true)
    const since = new Date()
    since.setDate(since.getDate() - days)
    const { data } = await supabase
      .from('completions')
      .select('*')
      .gte('completed_at', since.toISOString())
      .order('completed_at', { ascending: false })
    setLoading(false)
    if (data) setCompletions(data)
  }, [supabase])

  return { completions, loading, fetchCompletions }
}
