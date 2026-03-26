-- Enable RLS on all tables
alter table public.tasks       enable row level security;
alter table public.session     enable row level security;
alter table public.notes       enable row level security;
alter table public.completions enable row level security;
alter table public.settings    enable row level security;

-- tasks
create policy "Users can manage their own tasks"
  on public.tasks for all
  using (user_id = auth.uid())
  with check (user_id = auth.uid());

-- session
create policy "Users can manage their own session"
  on public.session for all
  using (user_id = auth.uid())
  with check (user_id = auth.uid());

-- notes
create policy "Users can manage their own notes"
  on public.notes for all
  using (user_id = auth.uid())
  with check (user_id = auth.uid());

-- completions
create policy "Users can manage their own completions"
  on public.completions for all
  using (user_id = auth.uid())
  with check (user_id = auth.uid());

-- settings
create policy "Users can manage their own settings"
  on public.settings for all
  using (user_id = auth.uid())
  with check (user_id = auth.uid());
