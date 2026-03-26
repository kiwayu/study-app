import { createBrowserClient } from '@supabase/ssr'

// TODO(Task 17): thread Database generic once web/lib/types/supabase.ts is generated.
// Until then, queries are typed as `any` which is acceptable during bootstrapping.
export function createClient() {
  return createBrowserClient(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!
  )
}
