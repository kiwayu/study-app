'use client'

import { createClient } from '@/lib/supabase/client'
import { Provider } from '@supabase/supabase-js'

const providers: { id: Provider; label: string; icon: string }[] = [
  {
    id: 'google',
    label: 'Continue with Google',
    icon: 'M21.35 11.1H12.18V13.83H18.69C18.36 17.64 15.19 19.27 12.19 19.27C8.36 19.27 5 16.28 5 12C5 7.73 8.54 4.73 12.19 4.73C15.19 4.73 17.08 6.7 17.08 6.7L19 4.72C19 4.72 16.56 2 12.1 2C6.42 2 2.03 6.8 2.03 12C2.03 17.05 6.16 22 12.19 22C17.6 22 21.5 18.33 21.5 12.33C21.5 11.76 21.35 11.1 21.35 11.1V11.1Z',
  },
  {
    id: 'github',
    label: 'Continue with GitHub',
    icon: 'M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844a9.59 9.59 0 012.504.337c1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z',
  },
  {
    id: 'apple',
    label: 'Sign in with Apple',
    icon: 'M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.8-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11z',
  },
]

export default function LoginPage() {
  const supabase = createClient()

  async function signInWith(provider: Provider) {
    const { error } = await supabase.auth.signInWithOAuth({
      provider,
      options: {
        redirectTo: `${window.location.origin}/auth/callback`,
      },
    })
    if (error) {
      console.error('OAuth sign-in error:', error.message)
    }
  }

  return (
    <div className="w-full max-w-sm space-y-8 px-4">
      {/* Header */}
      <div className="text-center space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight text-zinc-50">
          Study App
        </h1>
        <p className="text-sm text-zinc-400">
          Focus. Track. Improve.
        </p>
      </div>

      {/* Card */}
      <div className="rounded-2xl border border-zinc-800 bg-zinc-900/60 backdrop-blur-sm p-6 space-y-3 shadow-2xl">
        {providers.map((p) => (
          <button
            key={p.id}
            onClick={() => signInWith(p.id)}
            className="w-full flex items-center justify-center gap-3 rounded-full bg-zinc-800 hover:bg-zinc-700 active:bg-zinc-600 text-zinc-100 text-sm font-medium py-3 px-4 transition-colors duration-150"
          >
            <svg viewBox="0 0 24 24" className="w-5 h-5 fill-current flex-shrink-0" aria-hidden="true" focusable="false">
              <path d={p.icon} />
            </svg>
            {p.label}
          </button>
        ))}
      </div>

      <p className="text-center text-xs text-zinc-600">
        By signing in you agree to our terms of service.
      </p>
    </div>
  )
}
