'use client'

import type { Settings } from '@/lib/supabase/types'

interface Props { settings: Settings; onChange: <K extends keyof Settings>(k: K, v: Settings[K]) => void }

export function LayoutSettings({ settings, onChange }: Props) {
  return (
    <div className="space-y-4">
      <h3 className="text-xs font-semibold uppercase tracking-widest text-[var(--color-muted)]">Layout</h3>

      <div className="flex items-center justify-between">
        <span className="text-sm text-zinc-300">Sidebar position</span>
        <div className="flex gap-2">
          {(['left','right'] as const).map(pos => (
            <button key={pos} onClick={() => onChange('sidebar_position', pos)}
              aria-pressed={settings.sidebar_position === pos}
              className={`rounded-lg px-3 py-1.5 text-xs font-medium capitalize transition-colors ${
                settings.sidebar_position === pos
                  ? 'bg-[var(--color-accent)] text-white'
                  : 'bg-[var(--color-surface-2)] text-zinc-400 hover:text-zinc-200'
              }`}
            >{pos}</button>
          ))}
        </div>
      </div>

      <div className="flex items-center justify-between">
        <span className="text-sm text-zinc-300">Density</span>
        <div className="flex gap-2">
          {(['compact','default','spacious'] as const).map(d => (
            <button key={d} onClick={() => onChange('panel_density', d)}
              aria-pressed={settings.panel_density === d}
              className={`rounded-lg px-2.5 py-1.5 text-xs font-medium capitalize transition-colors ${
                settings.panel_density === d
                  ? 'bg-[var(--color-accent)] text-white'
                  : 'bg-[var(--color-surface-2)] text-zinc-400 hover:text-zinc-200'
              }`}
            >{d}</button>
          ))}
        </div>
      </div>

      <div>
        <div className="flex items-center justify-between mb-1.5">
          <label htmlFor="sidebar-width" className="text-sm text-zinc-300">Sidebar width</label>
          <span className="text-xs text-[var(--color-muted)]">{settings.sidebar_width}px</span>
        </div>
        <input id="sidebar-width" type="range" min={220} max={360} value={settings.sidebar_width}
          onChange={e => onChange('sidebar_width', parseInt(e.target.value, 10))}
          className="w-full accent-[var(--color-accent)]" />
      </div>
    </div>
  )
}
