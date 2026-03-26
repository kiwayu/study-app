import { createClient } from '@/lib/supabase/server'
import { signOut } from '@/app/actions/auth'

export default async function AppPage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  return (
    <main className="p-8">
      <h1 className="text-xl font-semibold">Welcome, {user?.email}</h1>
      <p className="text-zinc-400 mt-2">App features coming in Plan 2.</p>
      <form action={signOut} className="mt-4">
        <button type="submit" className="text-sm text-zinc-400 hover:text-zinc-200">
          Sign out
        </button>
      </form>
    </main>
  )
}
