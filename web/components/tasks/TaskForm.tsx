'use client'

import { useRef, useState } from 'react'
import type { Settings, Task } from '@/lib/supabase/types'

interface TaskFormProps {
  settings: Settings
  onAdd: (payload: Omit<Task, 'id' | 'user_id' | 'created_at' | 'pomodoros_done' | 'completed' | 'completed_at'>) => Promise<{ error?: string }>
}

export function TaskForm({ settings, onAdd }: TaskFormProps) {
  const titleRef = useRef<HTMLInputElement>(null)
  const [priority, setPriority] = useState(settings.default_priority)
  const [pomodoros, setPomodoros] = useState(String(settings.default_pomodoros))
  const [category, setCategory] = useState('')
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const title = titleRef.current?.value.trim() ?? ''
    if (!title) {
      setError('Title is required')
      titleRef.current?.focus()
      return
    }
    setError(null)

    const result = await onAdd({
      title,
      priority,
      category: category.trim() || null,
      pomodoros_est: Math.max(1, Math.min(9, parseInt(pomodoros, 10) || 1)),
      segment_mins: null,
      position: null,
    })

    if (result.error) {
      setError(result.error)
    } else {
      setError(null)
      if (titleRef.current) { titleRef.current.value = ''; titleRef.current.focus() }
      setPomodoros(String(settings.default_pomodoros))
      setPriority(settings.default_priority)
      setCategory('')
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-2" noValidate>
      {/* Title input */}
      <div>
        <label htmlFor="new-task-title" className="sr-only">Task title</label>
        <input
          id="new-task-title"
          ref={titleRef}
          type="text"
          placeholder="Add a task…"
          maxLength={200}
          className="w-full rounded-xl bg-[var(--color-surface)] border border-[var(--color-border)] focus:border-[var(--color-accent)] text-sm text-zinc-50 placeholder:text-[var(--color-muted)] px-4 py-2.5 outline-none transition-colors"
        />
        {error && <p className="text-xs text-[var(--color-danger)] mt-1">{error}</p>}
      </div>

      {/* Metadata row */}
      <div className="flex items-center gap-2">
        {/* Pomodoros */}
        <div className="flex items-center gap-1.5">
          <label htmlFor="new-task-pom" className="text-xs text-[var(--color-muted)]">🍅</label>
          <input
            id="new-task-pom"
            type="number"
            min={1} max={9}
            value={pomodoros}
            onChange={e => setPomodoros(e.target.value)}
            className="w-10 rounded-md bg-[var(--color-surface)] border border-[var(--color-border)] text-xs text-zinc-50 text-center px-1.5 py-1 outline-none focus:border-[var(--color-accent)]"
          />
        </div>

        {/* Priority */}
        <div>
          <label htmlFor="new-task-priority" className="sr-only">Priority</label>
          <select
            id="new-task-priority"
            value={priority}
            onChange={e => setPriority(e.target.value as Settings['default_priority'])}
            className="rounded-md bg-[var(--color-surface)] border border-[var(--color-border)] text-xs text-zinc-50 px-2 py-1 outline-none focus:border-[var(--color-accent)] appearance-none"
          >
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
          </select>
        </div>

        {/* Category */}
        <div className="flex-1">
          <label htmlFor="new-task-category" className="sr-only">Category</label>
          <input
            id="new-task-category"
            type="text"
            placeholder="Category"
            value={category}
            onChange={e => setCategory(e.target.value)}
            list="category-presets"
            className="w-full rounded-md bg-[var(--color-surface)] border border-[var(--color-border)] text-xs text-zinc-50 placeholder:text-[var(--color-muted)] px-2 py-1 outline-none focus:border-[var(--color-accent)]"
          />
          <datalist id="category-presets">
            {settings.category_presets.map(p => <option key={p} value={p} />)}
          </datalist>
        </div>

        {/* Submit */}
        <button
          type="submit"
          className="rounded-full bg-[var(--color-accent)] hover:bg-[var(--color-accent-dim)] text-white text-xs font-semibold px-3.5 py-1.5 transition-colors"
        >
          Add
        </button>
      </div>
    </form>
  )
}
