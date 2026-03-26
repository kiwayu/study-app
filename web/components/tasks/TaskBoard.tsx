'use client'

import { useState } from 'react'
import { TaskCard } from './TaskCard'
import { TaskForm } from './TaskForm'
import type { Task, Settings } from '@/lib/supabase/types'

interface TaskBoardProps {
  tasks: Task[]
  settings: Settings
  onAdd: Parameters<typeof TaskForm>[0]['onAdd']
  onUpdate: (id: string, changes: Partial<Task>) => void
  onDelete: (id: string) => void
  onComplete: (id: string) => void
  onStatsOpen: () => void
}

export function TaskBoard({ tasks, settings, onAdd, onUpdate, onDelete, onComplete, onStatsOpen }: TaskBoardProps) {
  const [filter, setFilter] = useState('')
  const [showCompleted, setShowCompleted] = useState(false)

  const filtered = tasks.filter(t => {
    if (!showCompleted && t.completed) return false
    if (!filter) return true
    return t.title.toLowerCase().includes(filter.toLowerCase()) ||
           (t.category ?? '').toLowerCase().includes(filter.toLowerCase())
  })

  return (
    <main className="flex-1 flex flex-col min-h-0 p-6 gap-4 overflow-y-auto">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-base font-semibold tracking-tight text-zinc-50">Tasks</h2>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowCompleted(v => !v)}
            aria-pressed={showCompleted}
            className="text-xs text-[var(--color-muted)] hover:text-zinc-200 transition-colors"
          >
            {showCompleted ? 'Hide done' : 'Show done'}
          </button>
          <button
            onClick={onStatsOpen}
            aria-label="Open stats"
            className="p-1.5 rounded-lg text-[var(--color-muted)] hover:text-zinc-200 hover:bg-[var(--color-surface-2)] transition-colors"
          >
            <svg viewBox="0 0 20 20" fill="currentColor" className="w-4 h-4" aria-hidden="true">
              <path d="M2 11a1 1 0 011-1h2a1 1 0 011 1v5a1 1 0 01-1 1H3a1 1 0 01-1-1v-5zm6-4a1 1 0 011-1h2a1 1 0 011 1v9a1 1 0 01-1 1H9a1 1 0 01-1-1V7zm6-3a1 1 0 011-1h2a1 1 0 011 1v12a1 1 0 01-1 1h-2a1 1 0 01-1-1V4z" />
            </svg>
          </button>
        </div>
      </div>

      {/* Search filter */}
      <div>
        <label htmlFor="task-filter" className="sr-only">Filter tasks</label>
        <input
          id="task-filter"
          type="text"
          placeholder="Filter tasks…"
          value={filter}
          onChange={e => setFilter(e.target.value)}
          className="w-full rounded-xl bg-[var(--color-surface)] border border-[var(--color-border)] focus:border-[var(--color-accent)] text-sm text-zinc-50 placeholder:text-[var(--color-muted)] px-4 py-2 outline-none transition-colors"
        />
      </div>

      {/* Add task form */}
      <TaskForm settings={settings} onAdd={onAdd} />

      {/* Task list */}
      <div className="flex flex-col gap-2">
        {filtered.length === 0 ? (
          <p className="text-sm text-[var(--color-muted)] text-center py-8">
            {filter ? 'No tasks match your filter.' : 'No tasks yet. Add one above.'}
          </p>
        ) : (
          filtered.map(task => (
            <TaskCard
              key={task.id}
              task={task}
              onUpdate={onUpdate}
              onDelete={onDelete}
              onComplete={onComplete}
            />
          ))
        )}
      </div>
    </main>
  )
}
