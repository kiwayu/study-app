'use client'

import { useState } from 'react'
import type { Task, Settings } from '@/lib/supabase/types'

interface TaskCardProps {
  task: Task
  settings: Settings
  onUpdate: (id: string, changes: Partial<Task>) => void
  onDelete: (id: string) => void
  onComplete: (id: string) => void
}

const PRIORITY_DOT: Record<string, string> = {
  high:   'bg-red-400',
  medium: 'bg-orange-400',
  low:    'bg-zinc-500',
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function TaskCard({ task, settings: _settings, onUpdate, onDelete, onComplete }: TaskCardProps) {
  const [editing, setEditing] = useState(false)
  const [editTitle, setEditTitle] = useState(task.title)

  function handleTitleBlur() {
    const trimmed = editTitle.trim()
    if (trimmed && trimmed !== task.title) {
      onUpdate(task.id, { title: trimmed })
    } else {
      setEditTitle(task.title)
    }
    setEditing(false)
  }

  return (
    <div
      className={`group flex items-start gap-3 rounded-xl bg-[var(--color-surface)] hover:bg-[var(--color-surface-2)] border border-[var(--color-border)] px-4 py-3 transition-all duration-150 ${
        task.completed ? 'opacity-50' : ''
      }`}
    >
      {/* Completion checkbox */}
      <button
        onClick={() => !task.completed && onComplete(task.id)}
        aria-label={task.completed ? 'Completed' : 'Mark complete'}
        className={`mt-0.5 w-4 h-4 shrink-0 rounded-full border-2 transition-colors ${
          task.completed
            ? 'bg-[var(--color-accent)] border-[var(--color-accent)]'
            : 'border-[var(--color-border)] hover:border-[var(--color-accent)]'
        }`}
      />

      {/* Main content */}
      <div className="flex-1 min-w-0">
        {editing ? (
          <input
            autoFocus
            value={editTitle}
            onChange={e => setEditTitle(e.target.value)}
            onBlur={handleTitleBlur}
            onKeyDown={e => e.key === 'Enter' && handleTitleBlur()}
            className="w-full bg-transparent text-sm text-zinc-50 focus:outline-none"
          />
        ) : (
          <button
            onClick={() => !task.completed && setEditing(true)}
            className={`text-left text-sm w-full ${
              task.completed ? 'line-through text-[var(--color-muted)]' : 'text-zinc-50'
            }`}
          >
            {task.title}
          </button>
        )}

        {/* Metadata row */}
        <div className="flex items-center gap-2 mt-1">
          {/* Priority dot */}
          <span className={`w-1.5 h-1.5 rounded-full ${PRIORITY_DOT[task.priority] ?? 'bg-zinc-500'}`} />

          {/* Pomodoro count */}
          <span className="text-xs text-[var(--color-muted)]">
            {task.pomodoros_done} / {task.pomodoros_est} 🍅
          </span>

          {/* Category */}
          {task.category && (
            <span className="text-xs text-[var(--color-muted)] truncate">
              {task.category}
            </span>
          )}
        </div>
      </div>

      {/* Delete button — visible on hover */}
      <button
        onClick={() => onDelete(task.id)}
        aria-label="Delete task"
        className="opacity-0 group-hover:opacity-100 p-1 rounded text-[var(--color-muted)] hover:text-[var(--color-danger)] transition-all"
      >
        <svg viewBox="0 0 16 16" fill="currentColor" className="w-3.5 h-3.5" aria-hidden="true">
          <path d="M5.5 5.5A.5.5 0 016 6v6a.5.5 0 01-1 0V6a.5.5 0 01.5-.5zm2.5 0a.5.5 0 01.5.5v6a.5.5 0 01-1 0V6a.5.5 0 01.5-.5zm3 .5a.5.5 0 00-1 0v6a.5.5 0 001 0V6z"/>
          <path fillRule="evenodd" d="M14.5 3a1 1 0 01-1 1H13v9a2 2 0 01-2 2H5a2 2 0 01-2-2V4h-.5a1 1 0 010-2h4a1 1 0 011-1h2a1 1 0 011 1h4a1 1 0 011 1zM4.118 4L4 4.059V13a1 1 0 001 1h6a1 1 0 001-1V4.059L11.882 4H4.118z" clipRule="evenodd"/>
        </svg>
      </button>
    </div>
  )
}
