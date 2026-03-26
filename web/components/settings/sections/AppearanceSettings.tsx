import type { Settings } from '@/lib/supabase/types'

interface Props { settings: Settings; onChange: (k: keyof Settings, v: unknown) => void }

const THEMES = [
  { id: 'serika-dark',  name: 'Serika Dark',  dark: true,  hex: '#e2b714' },
  { id: 'nord',         name: 'Nord',         dark: true,  hex: '#88c0d0' },
  { id: 'dracula',      name: 'Dracula',      dark: true,  hex: '#bd93f9' },
  { id: 'botanical',    name: 'Botanical',    dark: false, hex: '#7C9E87' },
  { id: 'rose-pine',    name: 'Rosé Pine',   dark: true,  hex: '#ebbcba' },
  { id: 'catppuccin',   name: 'Catppuccin',  dark: true,  hex: '#cba6f7' },
  { id: 'tokyo-night',  name: 'Tokyo Night', dark: true,  hex: '#7aa2f7' },
  { id: 'gruvbox',      name: 'Gruvbox',     dark: true,  hex: '#fabd2f' },
  { id: 'system',       name: 'System',      dark: true,  hex: '#71717a' },
]

export function AppearanceSettings({ settings, onChange }: Props) {
  return (
    <div className="space-y-4">
      <h3 className="text-xs font-semibold uppercase tracking-widest text-[var(--color-muted)]">Appearance</h3>

      <div>
        <p className="text-sm text-zinc-300 mb-2">Mode</p>
        <div className="flex gap-2">
          {(['system','dark','light'] as const).map(m => (
            <button
              key={m}
              onClick={() => onChange('dark_mode', m)}
              className={`flex-1 rounded-lg py-2 text-xs font-medium capitalize transition-colors ${
                settings.dark_mode === m
                  ? 'bg-[var(--color-accent)] text-white'
                  : 'bg-[var(--color-surface-2)] text-zinc-400 hover:text-zinc-200'
              }`}
            >
              {m}
            </button>
          ))}
        </div>
      </div>

      <div>
        <p className="text-sm text-zinc-300 mb-2">Theme</p>
        <div className="grid grid-cols-4 gap-2">
          {THEMES.map(t => (
            <button
              key={t.id}
              onClick={() => onChange('theme_id', t.id)}
              title={t.name}
              className={`relative aspect-square rounded-xl transition-all ${
                settings.theme_id === t.id ? 'ring-2 ring-[var(--color-accent)] ring-offset-2 ring-offset-[var(--color-bg)]' : ''
              }`}
              style={{ background: t.dark ? '#18181b' : '#f4f4f5' }}
            >
              <span className="absolute inset-0 m-auto w-3 h-3 rounded-full" style={{ background: t.hex }} />
            </button>
          ))}
        </div>
      </div>

      <div className="flex items-center justify-between">
        <label htmlFor="accent-color" className="text-sm text-zinc-300">Accent colour</label>
        <input
          id="accent-color"
          type="color"
          value={settings.color_accent}
          onChange={e => onChange('color_accent', e.target.value)}
          className="w-9 h-9 rounded-lg cursor-pointer bg-transparent border border-[var(--color-border)] p-0.5"
        />
      </div>
    </div>
  )
}
