'use client'

import { useState } from 'react'
import type { Settings, Priority } from '@/lib/supabase/types'

interface Props { settings: Settings; onChange: <K extends keyof Settings>(k: K, v: Settings[K]) => void }

export function TaskDefaultsSettings({ settings, onChange }: Props) {
  const [newPreset, setNewPreset] = useState('')

  function addPreset() {
    const trimmed = newPreset.trim()
    if (!trimmed || settings.category_presets.includes(trimmed)) return
    onChange('category_presets', [...settings.category_presets, trimmed])
    setNewPreset('')
  }

  function removePreset(p: string) {
    onChange('category_presets', settings.category_presets.filter(x => x !== p))
  }

  return (
    <div className="space-y-4">
      <h3 className="text-xs font-semibold uppercase tracking-widest text-[var(--color-muted)]">Task Defaults</h3>

      <div className="flex items-center justify-between">
        <label htmlFor="def-priority" className="text-sm text-zinc-300">Default priority</label>
        <select id="def-priority" value={settings.default_priority}
          onChange={e => onChange('default_priority', e.target.value as Priority)}
          className="rounded-lg bg-[var(--color-surface-2)] border border-[var(--color-border)] text-sm text-zinc-50 px-3 py-1.5 outline-none focus:border-[var(--color-accent)]"
        >
          <option value="high">High</option>
          <option value="medium">Medium</option>
          <option value="low">Low</option>
        </select>
      </div>

      <div className="flex items-center justify-between">
        <label htmlFor="def-pom" className="text-sm text-zinc-300">Default pomodoros</label>
        <input id="def-pom" type="number" min={1} max={9} value={settings.default_pomodoros}
          onChange={e => onChange('default_pomodoros', Math.max(1, Math.min(9, parseInt(e.target.value, 10) || 1)))}
          className="w-20 rounded-lg bg-[var(--color-surface-2)] border border-[var(--color-border)] text-sm text-zinc-50 text-center px-2 py-1.5 outline-none focus:border-[var(--color-accent)]"
        />
      </div>

      <div>
        <p className="text-sm text-zinc-300 mb-2">Category presets</p>
        <div className="flex flex-wrap gap-1.5 mb-2">
          {settings.category_presets.map(p => (
            <span key={p} className="flex items-center gap-1 rounded-full bg-[var(--color-surface-2)] text-xs text-zinc-300 px-2.5 py-1">
              {p}
              <button onClick={() => removePreset(p)} aria-label={`Remove ${p}`} className="text-[var(--color-muted)] hover:text-[var(--color-danger)] ml-0.5">×</button>
            </span>
          ))}
        </div>
        <div className="flex gap-2">
          <input aria-label="New category name" value={newPreset} onChange={e => setNewPreset(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && addPreset()}
            placeholder="New category…"
            className="flex-1 rounded-lg bg-[var(--color-surface-2)] border border-[var(--color-border)] text-sm text-zinc-50 placeholder:text-[var(--color-muted)] px-3 py-1.5 outline-none focus:border-[var(--color-accent)]"
          />
          <button onClick={addPreset}
            className="rounded-lg bg-[var(--color-surface-2)] hover:bg-zinc-600 text-zinc-300 text-sm px-3 py-1.5 transition-colors"
          >Add</button>
        </div>
      </div>
    </div>
  )
}
