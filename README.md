# Study Session — Pomodoro Manager

A floating, phone-screen-style Pomodoro timer. Go backend + vanilla JS frontend.

## Run

```
go run .
```

Open [http://localhost:8080](http://localhost:8080)

Requires Go 1.22+. No other dependencies.

State persists to `data/state.json` on every write.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/tasks` | List all tasks (sorted by order) |
| `POST` | `/api/tasks` | Create task → returns `Task` (201) |
| `PUT` | `/api/tasks/{id}` | Update task fields (partial) → returns `Task` |
| `DELETE` | `/api/tasks/{id}` | Delete task → 204 |
| `GET` | `/api/settings` | Get settings |
| `PUT` | `/api/settings` | Replace settings → returns `Settings` |
| `GET` | `/api/session` | Get session state |
| `POST` | `/api/session/start` | Start/advance segment → returns `SessionState` |
| `POST` | `/api/session/pause` | Pause timer → returns `SessionState` |
| `POST` | `/api/session/stop` | Full reset → returns `SessionState` |
| `PUT` | `/api/session/totals` | Update totalElapsed and toast timestamps |

## Palette Customisation

Edit the CSS custom properties in `static/css/base.css`:

```css
--color-accent: #7C9E87;      /* Sage — timer ring, buttons, focus ring */
--color-accent-dim: #5a7a65;  /* Sage hover state */
--color-bg: #09090b;          /* Page background */
--color-surface: #18181b;     /* Card background */
--color-border: #27272a;      /* Borders, dividers */
--color-text: #f4f4f5;        /* Primary text */
--color-muted: #a1a1aa;       /* Secondary text, icons */
```

To change the accent to indigo, for example:
```css
--color-accent: #818cf8;
--color-accent-dim: #6366f1;
```

## Segment Sequence

The 8-slot Pomodoro cycle (repeats on wrap):

```
Focus → Short Break → Focus → Short Break →
Focus → Short Break → Focus → Long Break → (wrap)
```

Default durations: Focus 25 min · Short Break 5 min · Long Break 15 min.
All durations configurable in Settings.
