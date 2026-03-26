export type Priority = 'high' | 'medium' | 'low'
export type SessionStatus = 'idle' | 'running' | 'paused'
export type SegmentType = 'focus' | 'short_break' | 'long_break'
export type DarkMode = 'system' | 'dark' | 'light'
export type SoundPack = 'mechanical' | 'soft' | 'minimal'
export type NotificationStyle = 'toast' | 'native' | 'silent'
export type Density = 'compact' | 'default' | 'spacious'

export interface Task {
  id: string
  user_id: string
  title: string
  priority: Priority
  category: string | null
  pomodoros_est: number
  pomodoros_done: number
  segment_mins: number | null
  completed: boolean
  completed_at: string | null
  position: number | null
  created_at: string
}

export interface Session {
  id: string
  user_id: string
  status: SessionStatus
  segment_type: SegmentType
  segment_index: number
  started_at: string | null
  banked_ms: number
  pomodoro_count: number
  totals: {
    totalElapsedMs: number
    lastWaterAt: number
    lastStretchAt: number
  }
}

export interface Note {
  id: string
  user_id: string
  date: string
  text: string | null
  updated_at: string
}

export interface Completion {
  id: string
  user_id: string
  task_id: string | null
  completed_at: string
  pomodoros_est: number | null
  pomodoros_actual: number | null
}

export interface Settings {
  id: string
  user_id: string
  pomodoro_duration: number
  short_break: number
  long_break: number
  water_interval: number
  stretch_interval: number
  theme_id: string
  color_accent: string
  dark_mode: DarkMode
  font_family: string
  font_size: number
  line_spacing: number
  sidebar_position: 'left' | 'right'
  panel_density: Density
  sidebar_width: number
  sound_enabled: boolean
  sound_volume: number
  sound_pack: SoundPack
  notification_style: NotificationStyle
  default_priority: Priority
  default_pomodoros: number
  category_presets: string[]
  updated_at: string
}
