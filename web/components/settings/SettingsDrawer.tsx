'use client'

import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { createClient } from '@/lib/supabase/client'
import { TimerSettings } from './sections/TimerSettings'
import { AppearanceSettings } from './sections/AppearanceSettings'
import { TypographySettings } from './sections/TypographySettings'
import { LayoutSettings } from './sections/LayoutSettings'
import { SoundSettings } from './sections/SoundSettings'
import { NotificationSettings } from './sections/NotificationSettings'
import { TaskDefaultsSettings } from './sections/TaskDefaultsSettings'
import type { Settings } from '@/lib/supabase/types'

interface SettingsDrawerProps {
  initialSettings: Settings
  isOpen: boolean
  onClose: () => void
  onSave: (settings: Settings) => void
}

export function SettingsDrawer({ initialSettings, isOpen, onClose, onSave }: SettingsDrawerProps) {
  const [draft, setDraft] = useState<Settings>(initialSettings)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const firstInputRef = useRef<HTMLDivElement>(null)
  const supabase = useMemo(() => createClient(), [])

  // Reset draft when drawer opens
  useEffect(() => {
    if (isOpen) {
      setDraft(initialSettings)
      setError(null)
      setTimeout(() => (firstInputRef.current?.querySelector('input,select,button') as HTMLElement | null)?.focus(), 50)
    }
  }, [isOpen, initialSettings])

  // Escape closes
  useEffect(() => {
    if (!isOpen) return
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [isOpen, onClose])

  const handleChange = useCallback((key: keyof Settings, value: unknown) => {
    setDraft(d => ({ ...d, [key]: value }))
  }, [])

  async function handleSave() {
    setSaving(true)
    setError(null)
    const { error } = await supabase
      .from('settings')
      .upsert({ ...draft, updated_at: new Date().toISOString() })
    setSaving(false)
    if (error) { setError(error.message); return }
    onSave(draft)
    onClose()
  }

  return (
    <>
      {/* Backdrop */}
      {isOpen && (
        <div
          className="fixed inset-0 bg-black/50 backdrop-blur-sm z-[var(--z-drawer)]"
          onClick={onClose}
          aria-hidden="true"
        />
      )}

      {/* Drawer */}
      <div
        role="dialog"
        aria-label="Settings"
        aria-modal="true"
        className={`fixed top-0 right-0 h-full w-80 bg-[var(--color-surface)] border-l border-[var(--color-border)] shadow-lg z-[calc(var(--z-drawer)+1)] flex flex-col transition-transform duration-300 ease-spring ${
          isOpen ? 'translate-x-0' : 'translate-x-full'
        }`}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-5 py-4 border-b border-[var(--color-border)]">
          <h2 className="text-sm font-semibold tracking-tight text-zinc-50">Settings</h2>
          <button onClick={onClose} aria-label="Close settings" className="p-1.5 rounded-lg text-[var(--color-muted)] hover:text-zinc-200 hover:bg-[var(--color-surface-2)] transition-colors">
            <svg viewBox="0 0 20 20" fill="currentColor" className="w-4 h-4" aria-hidden="true"><path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" /></svg>
          </button>
        </div>

        {/* Scrollable content */}
        <div ref={firstInputRef} className="flex-1 overflow-y-auto px-5 py-4 space-y-8">
          <TimerSettings settings={draft} onChange={handleChange} />
          <AppearanceSettings settings={draft} onChange={handleChange} />
          <TypographySettings settings={draft} onChange={handleChange} />
          <LayoutSettings settings={draft} onChange={handleChange} />
          <SoundSettings settings={draft} onChange={handleChange} />
          <NotificationSettings settings={draft} onChange={handleChange} />
          <TaskDefaultsSettings settings={draft} onChange={handleChange} />
        </div>

        {/* Footer */}
        <div className="px-5 py-4 border-t border-[var(--color-border)] space-y-2">
          {error && <p className="text-xs text-[var(--color-danger)]">{error}</p>}
          <div className="flex gap-2">
            <button onClick={onClose} className="flex-1 rounded-full bg-[var(--color-surface-2)] hover:bg-zinc-600 text-zinc-300 text-sm font-medium py-2.5 transition-colors">Cancel</button>
            <button onClick={handleSave} disabled={saving} className="flex-1 rounded-full bg-[var(--color-accent)] hover:bg-[var(--color-accent-dim)] disabled:opacity-50 text-white text-sm font-semibold py-2.5 transition-colors">
              {saving ? 'Saving…' : 'Save'}
            </button>
          </div>
        </div>
      </div>
    </>
  )
}
