'use client'

import { useEffect, useState } from 'react'
import { Heatmap } from './Heatmap'
import { useStats } from '@/lib/hooks/useStats'

interface StatsPanelProps {
  isOpen: boolean
  onClose: () => void
}

type Tab = 'year' | 'month' | 'estimation'

export function StatsPanel({ isOpen, onClose }: StatsPanelProps) {
  const [tab, setTab] = useState<Tab>('year')
  const { completions, loading, fetchCompletions } = useStats()

  useEffect(() => {
    if (isOpen) fetchCompletions(365)
  }, [isOpen, fetchCompletions])

  useEffect(() => {
    if (!isOpen) return
    function onKey(e: KeyboardEvent) { if (e.key === 'Escape') onClose() }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [isOpen, onClose])

  const completionsByDate = completions.reduce<Record<string, number>>((acc, c) => {
    const key = c.completed_at.slice(0, 10)
    acc[key] = (acc[key] ?? 0) + 1
    return acc
  }, {})

  return (
    <>
      {isOpen && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-[var(--z-modal)]" onClick={onClose} aria-hidden="true" />
      )}

      <div
        role="dialog"
        aria-label="Statistics"
        aria-modal="true"
        className={`fixed inset-x-0 bottom-0 max-h-[80vh] rounded-t-2xl bg-[var(--color-surface)] border-t border-[var(--color-border)] shadow-lg z-[calc(var(--z-modal)+1)] flex flex-col transition-transform duration-300 ease-spring ${
          isOpen ? 'translate-y-0' : 'translate-y-full'
        }`}
      >
        {/* Handle */}
        <div className="flex justify-center pt-3 pb-1">
          <div className="w-10 h-1 rounded-full bg-[var(--color-border)]" />
        </div>

        {/* Header + tabs */}
        <div className="flex items-center justify-between px-5 pb-3 border-b border-[var(--color-border)]">
          <div className="flex gap-1" role="tablist" aria-label="Stats view">
            {(['year','month','estimation'] as Tab[]).map(t => (
              <button key={t} onClick={() => setTab(t)}
                role="tab"
                aria-selected={tab === t}
                className={`rounded-lg px-3 py-1.5 text-xs font-medium capitalize transition-colors ${
                  tab === t ? 'bg-[var(--color-accent)] text-white' : 'text-[var(--color-muted)] hover:text-zinc-200'
                }`}
              >{t}</button>
            ))}
          </div>
          <button onClick={onClose} aria-label="Close stats" className="p-1.5 rounded-lg text-[var(--color-muted)] hover:text-zinc-200 transition-colors">
            <svg viewBox="0 0 20 20" fill="currentColor" className="w-4 h-4" aria-hidden="true" focusable="false"><path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" /></svg>
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto px-5 py-4">
          {loading && (
            <div className="space-y-2">
              {[...Array(3)].map((_,i) => (
                <div key={i} className="h-4 rounded bg-[var(--color-surface-2)] animate-pulse" />
              ))}
            </div>
          )}

          {!loading && tab === 'year' && (
            <div>
              <p className="text-xs text-[var(--color-muted)] mb-3">Last 52 weeks</p>
              <Heatmap completionsByDate={completionsByDate} />
            </div>
          )}

          {!loading && tab === 'month' && (() => {
            // Build current month calendar
            const now   = new Date()
            const year  = now.getFullYear()
            const month = now.getMonth()
            const daysInMonth = new Date(year, month + 1, 0).getDate()
            const days = Array.from({ length: daysInMonth }, (_, i) => {
              const d   = new Date(year, month, i + 1)
              const key = d.toLocaleDateString('sv-SE')
              return { day: i + 1, key, count: completionsByDate[key] ?? 0 }
            })
            return (
              <div>
                <p className="text-xs text-[var(--color-muted)] mb-3">
                  {now.toLocaleString('default', { month: 'long', year: 'numeric' })}
                </p>
                <div className="grid grid-cols-7 gap-1">
                  {days.map(({ day, key, count }) => (
                    <div key={key} title={`${key}: ${count} tasks`}
                      className={`aspect-square rounded-lg flex items-center justify-center text-xs font-medium transition-colors ${
                        count > 0 ? 'bg-[var(--color-accent)] text-white' : 'bg-[var(--color-surface-2)] text-[var(--color-muted)]'
                      }`}
                    >{day}</div>
                  ))}
                </div>
              </div>
            )
          })()}

          {!loading && tab === 'estimation' && (
            <div className="space-y-2">
              {completions.slice(0, 30).map(c => (
                <div key={c.id} className="flex items-center justify-between text-sm py-1.5 border-b border-[var(--color-border)]">
                  <span className="text-zinc-300 truncate mr-4 flex-1 text-xs font-mono">{c.task_id?.slice(0, 8) ?? '—'}</span>
                  <span className="text-[var(--color-muted)] shrink-0 text-xs">
                    {c.pomodoros_est ?? '?'} est → {c.pomodoros_actual ?? '?'} actual
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </>
  )
}
