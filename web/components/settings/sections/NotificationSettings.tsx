import type { Settings } from '@/lib/supabase/types'

interface Props { settings: Settings; onChange: (k: keyof Settings, v: unknown) => void }

export function NotificationSettings({ settings, onChange }: Props) {
  return (
    <div className="space-y-4">
      <h3 className="text-xs font-semibold uppercase tracking-widest text-[var(--color-muted)]">Notifications</h3>
      <div className="flex items-center justify-between">
        <span className="text-sm text-zinc-300">Style</span>
        <div className="flex gap-2">
          {(['toast','native','silent'] as const).map(s => (
            <button key={s} onClick={() => onChange('notification_style', s)}
              className={`rounded-lg px-3 py-1.5 text-xs font-medium capitalize transition-colors ${
                settings.notification_style === s
                  ? 'bg-[var(--color-accent)] text-white'
                  : 'bg-[var(--color-surface-2)] text-zinc-400 hover:text-zinc-200'
              }`}
            >{s}</button>
          ))}
        </div>
      </div>
    </div>
  )
}
