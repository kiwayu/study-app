# Study Session — Pomodoro Manager

A Pomodoro timer and task manager built with Go and vanilla JavaScript. Runs as a multi-user web app (PostgreSQL + OAuth) or a single-user desktop app (WebView2).

## Quick Start

### Web Mode (default)

Requires Go 1.22+, PostgreSQL 16+.

```bash
cp .env.example .env    # configure OAuth credentials, DB URL, JWT secret
go run .                # starts on :8080
```

Or with Docker:

```bash
docker-compose up       # app + PostgreSQL
```

Open [http://localhost:8080](http://localhost:8080) — sign in with Google or GitHub.

### Desktop Mode

No database or OAuth required. State persists to `data/state.json`.

```bash
go run . --desktop      # opens a 393×852 WebView2 window
```

## Features

- **Timer**: 8-slot Pomodoro cycle with configurable durations, pause/resume, progress ring
- **Tasks**: CRUD, drag-drop reorder, priority labels, categories, search/filter
- **Stats**: GitHub-style completion heatmap, estimation accuracy, daily session notes
- **Themes**: 22 color presets with live switching
- **Desktop**: Always-on-top, native toast notifications, title bar theming (Windows 11)
- **Auth**: Google + GitHub OAuth2, JWT access tokens, refresh token rotation
- **Security**: Rate limiting, CSRF protection, CSP headers, parameterized SQL

## Architecture

```
main.go              # Entry point — web or desktop mode
config/              # Environment-based configuration
models/              # Data structures (User, Task, Settings, Session, Note)
db/                  # PostgreSQL connection, migrations, repositories
auth/                # OAuth2 flows, JWT, auth middleware
handlers/            # HTTP handlers (tasks, settings, session, notes, stats, health)
middleware/          # CORS, CSRF, rate limiting, security headers, logging
static/              # Frontend — vanilla JS (ES6 modules), CSS, HTML
```

## API

All `/api/*` routes require authentication in web mode.

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/tasks` | List tasks (sorted by order) |
| `POST` | `/api/tasks` | Create task (201) |
| `PUT` | `/api/tasks/{id}` | Partial update |
| `DELETE` | `/api/tasks/{id}` | Delete (204) |
| `GET/PUT` | `/api/settings` | Get or replace settings |
| `GET` | `/api/session` | Current session state |
| `POST` | `/api/session/start` | Start segment |
| `POST` | `/api/session/pause` | Pause timer |
| `POST` | `/api/session/stop` | Stop and reset |
| `PUT` | `/api/session/totals` | Update elapsed/reminder timestamps |
| `GET` | `/api/stats/completions` | Completion counts by date |
| `GET` | `/api/stats/estimation` | Estimated vs actual pomodoros |
| `GET/PUT` | `/api/notes/{date}` | Daily session notes |
| `GET` | `/api/me` | Current user profile |
| `GET` | `/health` | Health check (no auth) |

### Auth Routes (no auth required)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/auth/login/google` | Redirect to Google OAuth |
| `GET` | `/auth/login/github` | Redirect to GitHub OAuth |
| `GET` | `/auth/callback/google` | Google OAuth callback |
| `GET` | `/auth/callback/github` | GitHub OAuth callback |
| `POST` | `/auth/refresh` | Refresh access token |
| `POST` | `/auth/logout` | Clear tokens |

## Configuration

See [`.env.example`](.env.example) for all environment variables. Key settings:

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | — | PostgreSQL connection string |
| `GOOGLE_CLIENT_ID` | — | Google OAuth client ID |
| `GITHUB_CLIENT_ID` | — | GitHub OAuth client ID |
| `JWT_SECRET` | — | Secret for signing JWTs |
| `PORT` | `8080` | Server port |
| `ENV` | `development` | `production` enforces required vars and secure cookies |

## Testing

```bash
go test -v ./...    # 21 tests across 4 packages
```

## Pomodoro Cycle

```
Focus → Short Break → Focus → Short Break →
Focus → Short Break → Focus → Long Break → (repeat)
```

Defaults: Focus 25 min, Short Break 5 min, Long Break 15 min. All configurable in Settings.
