import type { Settings } from '@/lib/supabase/types'

interface Props { settings: Settings; onChange: (k: keyof Settings, v: unknown) => void }

const FONTS = [
  { id: 'inter',          name: 'Inter' },
  { id: 'system',         name: 'System' },
  { id: 'jetbrains-mono', name: 'JetBrains Mono' },
  { id: 'geist',          name: 'Geist' },
]

export function TypographySettings({ settings, onChange }: Props) {
  return (
    <div className="space-y-4">
      <h3 className="text-xs font-semibold uppercase tracking-widest text-[var(--color-muted)]">Typography</h3>

      <div className="flex items-center justify-between">
        <label htmlFor="font-family" className="text-sm text-zinc-300">Font</label>
        <select
          id="font-family"
          value={settings.font_family}
          onChange={e => onChange('font_family', e.target.value)}
          className="rounded-lg bg-[var(--color-surface-2)] border border-[var(--color-border)] text-sm text-zinc-50 px-3 py-1.5 outline-none focus:border-[var(--color-accent)]"
        >
          {FONTS.map(f => <option key={f.id} value={f.id}>{f.name}</option>)}
        </select>
      </div>

      <div>
        <div className="flex items-center justify-between mb-1.5">
          <label htmlFor="font-size" className="text-sm text-zinc-300">Size</label>
          <span className="text-xs text-[var(--color-muted)]">{settings.font_size}px</span>
        </div>
        <input id="font-size" type="range" min={12} max={20} value={settings.font_size}
          onChange={e => onChange('font_size', parseInt(e.target.value, 10))}
          className="w-full accent-[var(--color-accent)]" />
      </div>

      <div>
        <div className="flex items-center justify-between mb-1.5">
          <label htmlFor="line-spacing" className="text-sm text-zinc-300">Line spacing</label>
          <span className="text-xs text-[var(--color-muted)]">{settings.line_spacing.toFixed(1)}</span>
        </div>
        <input id="line-spacing" type="range" min={1.2} max={2.0} step={0.1} value={settings.line_spacing}
          onChange={e => onChange('line_spacing', parseFloat(e.target.value))}
          className="w-full accent-[var(--color-accent)]" />
      </div>
    </div>
  )
}
