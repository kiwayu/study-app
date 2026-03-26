-- tasks
create table if not exists public.tasks (
  id             uuid primary key default gen_random_uuid(),
  user_id        uuid references auth.users not null,
  title          text not null,
  priority       text check (priority in ('high','medium','low')) default 'medium',
  category       text,
  pomodoros_est  int default 1,
  pomodoros_done int default 0,
  segment_mins   int,
  completed      boolean default false,
  completed_at   timestamptz,
  position       int,
  created_at     timestamptz default now()
);

-- session (one row per user — last-write-wins across devices)
create table if not exists public.session (
  id             uuid primary key default gen_random_uuid(),
  user_id        uuid references auth.users not null unique,
  status         text check (status in ('idle','running','paused')) default 'idle',
  segment_type   text check (segment_type in ('focus','short_break','long_break')) default 'focus',
  segment_index  int default 0,
  started_at     timestamptz,
  banked_ms      bigint default 0,
  pomodoro_count int default 0,
  -- totals shape: { totalElapsedMs: number, lastWaterAt: number, lastStretchAt: number }
  totals         jsonb default '{"totalElapsedMs":0,"lastWaterAt":0,"lastStretchAt":0}'
);

-- notes (one per user per calendar date)
create table if not exists public.notes (
  id         uuid primary key default gen_random_uuid(),
  user_id    uuid references auth.users not null,
  date       date not null,
  text       text,
  updated_at timestamptz default now(),
  unique (user_id, date)
);

-- completions (stats — task_id set null on task delete to preserve history)
create table if not exists public.completions (
  id               uuid primary key default gen_random_uuid(),
  user_id          uuid references auth.users not null,
  task_id          uuid references public.tasks on delete set null,
  completed_at     timestamptz default now(),
  pomodoros_est    int,
  pomodoros_actual int
);

-- settings (one row per user)
create table if not exists public.settings (
  id                  uuid primary key default gen_random_uuid(),
  user_id             uuid references auth.users not null unique,
  pomodoro_duration   int default 25,
  short_break         int default 5,
  long_break          int default 15,
  water_interval      int default 45,
  stretch_interval    int default 60,
  theme_id            text default 'system',
  color_accent        text default '#7C9E87',
  dark_mode           text default 'system',
  font_family         text default 'inter',
  font_size           int default 14,
  line_spacing        float default 1.5,
  sidebar_position    text default 'left',
  panel_density       text default 'default',
  sidebar_width       int default 268,
  sound_enabled       boolean default true,
  sound_volume        int default 70,
  sound_pack          text default 'soft',
  notification_style  text default 'toast',
  default_priority    text default 'medium',
  default_pomodoros   int default 1,
  category_presets    jsonb default '[]',
  updated_at          timestamptz default now()
);
