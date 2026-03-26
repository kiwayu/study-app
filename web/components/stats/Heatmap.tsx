interface HeatmapProps {
  completionsByDate: Record<string, number>  // { 'YYYY-MM-DD': count }
}

function dateKey(d: Date) {
  return d.toLocaleDateString('sv-SE') // 'YYYY-MM-DD' in local timezone
}

function getLevel(count: number): number {
  if (count === 0) return 0
  if (count <= 1) return 1
  if (count <= 3) return 2
  if (count <= 6) return 3
  return 4
}

const LEVEL_COLORS = [
  'bg-[var(--color-surface-2)]',
  'bg-[var(--color-accent)]/30',
  'bg-[var(--color-accent)]/50',
  'bg-[var(--color-accent)]/75',
  'bg-[var(--color-accent)]',
]

export function Heatmap({ completionsByDate }: HeatmapProps) {
  // Build 52-week grid
  const today = new Date()
  const weeks: Date[][] = []
  const cursor = new Date(today)
  cursor.setDate(cursor.getDate() - cursor.getDay()) // start of current week (Sunday)

  for (let w = 51; w >= 0; w--) {
    const week: Date[] = []
    for (let d = 0; d < 7; d++) {
      const day = new Date(cursor)
      day.setDate(cursor.getDate() - (w * 7) + d)
      week.push(day)
    }
    weeks.push(week)
  }

  return (
    <div className="flex gap-1 overflow-x-auto pb-1">
      {weeks.map((week, wi) => (
        <div key={wi} className="flex flex-col gap-1">
          {week.map((day) => {
            const key   = dateKey(day)
            const count = completionsByDate[key] ?? 0
            const level = getLevel(count)
            const isToday = dateKey(day) === dateKey(today)
            return (
              <div
                key={key}
                title={`${key}: ${count} task${count !== 1 ? 's' : ''}`}
                className={`w-3 h-3 rounded-sm transition-colors ${LEVEL_COLORS[level]} ${
                  isToday ? 'ring-1 ring-[var(--color-accent)]' : ''
                }`}
              />
            )
          })}
        </div>
      ))}
    </div>
  )
}
