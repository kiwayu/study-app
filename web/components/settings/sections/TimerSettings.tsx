'use client'

import type { Settings } from '@/lib/supabase/types'

interface Props { settings: Settings; onChange: <K extends keyof Settings>(k: K, v: Settings[K]) => void }

function Field({ label, id, children }: { label: string; id: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-4">
      <label htmlFor={id} className="text-sm text-zinc-300 shrink-0">{label}</label>
      {children}
    </div>
  )
}

function NumInput({ id, value, min, max, onChange }: { id: string; value: number; min: number; max: number; onChange: (v: number) => void }) {
  return (
    <input
      id={id} type="number" min={min} max={max} value={value}
      onChange={e => onChange(Math.max(min, Math.min(max, parseInt(e.target.value, 10) || min)))}
      className="w-20 rounded-lg bg-[var(--color-surface-2)] border border-[var(--color-border)] text-sm text-zinc-50 text-center px-2 py-1.5 outline-none focus:border-[var(--color-accent)]"
    />
  )
}

export function TimerSettings({ settings, onChange }: Props) {
  return (
    <div className="space-y-4">
      <h3 className="text-xs font-semibold uppercase tracking-widest text-[var(--color-muted)]">Timer</h3>
      <Field label="Focus (min)" id="pom-dur"><NumInput id="pom-dur" value={settings.pomodoro_duration} min={1} max={120} onChange={v => onChange('pomodoro_duration', v)} /></Field>
      <Field label="Short break (min)" id="short-brk"><NumInput id="short-brk" value={settings.short_break} min={1} max={60} onChange={v => onChange('short_break', v)} /></Field>
      <Field label="Long break (min)" id="long-brk"><NumInput id="long-brk" value={settings.long_break} min={1} max={120} onChange={v => onChange('long_break', v)} /></Field>
      <Field label="Water reminder (min)" id="water"><NumInput id="water" value={settings.water_interval} min={5} max={180} onChange={v => onChange('water_interval', v)} /></Field>
      <Field label="Stretch reminder (min)" id="stretch"><NumInput id="stretch" value={settings.stretch_interval} min={5} max={180} onChange={v => onChange('stretch_interval', v)} /></Field>
    </div>
  )
}
