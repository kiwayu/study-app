'use client'

import type { Settings, SoundPack } from '@/lib/supabase/types'

interface Props { settings: Settings; onChange: <K extends keyof Settings>(k: K, v: Settings[K]) => void }

export function SoundSettings({ settings, onChange }: Props) {
  return (
    <div className="space-y-4">
      <h3 className="text-xs font-semibold uppercase tracking-widest text-[var(--color-muted)]">Sounds</h3>

      <div className="flex items-center justify-between">
        <label htmlFor="sound-toggle" className="text-sm text-zinc-300">Enable sounds</label>
        <button
          id="sound-toggle"
          role="switch"
          aria-checked={settings.sound_enabled}
          onClick={() => onChange('sound_enabled', !settings.sound_enabled)}
          className={`relative w-10 h-6 rounded-full transition-colors ${settings.sound_enabled ? 'bg-[var(--color-accent)]' : 'bg-[var(--color-surface-2)]'}`}
        >
          <span className={`absolute top-1 w-4 h-4 rounded-full bg-white shadow transition-transform ${settings.sound_enabled ? 'left-5' : 'left-1'}`} />
        </button>
      </div>

      {settings.sound_enabled && (
        <>
          <div>
            <div className="flex items-center justify-between mb-1.5">
              <label htmlFor="sound-vol" className="text-sm text-zinc-300">Volume</label>
              <span className="text-xs text-[var(--color-muted)]">{settings.sound_volume}%</span>
            </div>
            <input id="sound-vol" type="range" min={0} max={100} value={settings.sound_volume}
              onChange={e => onChange('sound_volume', parseInt(e.target.value, 10))}
              className="w-full accent-[var(--color-accent)]" />
          </div>

          <div className="flex items-center justify-between">
            <label htmlFor="sound-pack" className="text-sm text-zinc-300">Sound pack</label>
            <select id="sound-pack" value={settings.sound_pack}
              onChange={e => onChange('sound_pack', e.target.value as SoundPack)}
              className="rounded-lg bg-[var(--color-surface-2)] border border-[var(--color-border)] text-sm text-zinc-50 px-3 py-1.5 outline-none focus:border-[var(--color-accent)]"
            >
              <option value="mechanical">Mechanical</option>
              <option value="soft">Soft</option>
              <option value="minimal">Minimal</option>
            </select>
          </div>
        </>
      )}
    </div>
  )
}
