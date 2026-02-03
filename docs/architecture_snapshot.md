# Living Architecture Snapshot

## Purpose
Provide a single reference for the current Connect4 architecture that stays in sync with the code. This page is meant to help interviewers, contributors, and new teammates quickly grasp the system boundaries, tooling, and how the pieces fit together.

## System Overview
- **Backend**: Go 1.23 + Gin handles HTTP endpoints and orchestrates the WebSocket hub, matchmaking, and game service layers.
- **Realtime layer**: Gorilla WebSocket connections are managed by `internal/websocket/hub.go` which keeps `gameRooms` and per-connection state in sync.
- **Game logic**: `internal/game/service.go` enforces session lifecycle, board validation, and analytics hooks while `internal/game/engine.go` keeps the board consistent.
- **AI**: `internal/bot/minimax.go` drives the bot opponent using alpha-beta pruning and iterative deepening.
- **Analytics**: Events produced in `internal/analytics/producer.go` are streamed to Kafka (Confluent Cloud) and consumed by `cmd/analytics/main.go` for metrics.
- **Frontend**: React + Vite in `web/` connects via `web/src/services/websocket.ts` for real-time messaging and uses contexts (`AuthContext`, `PlayerContext`) to surface user and game state.

## Deployment Diagram (Current)
1. **Render.com (Backend)**
   - Docker container defined in `Dockerfile.server`
   - Starts `cmd/server/main.go`
   - Connects to Supabase PostgreSQL and, if configured, Kafka's `game-events` topic
2. **Vercel (Frontend)**
   - Deploys `web/` with Vite and connects to Render backend via REST/WebSocket endpoints
3. **Supabase PostgreSQL**
   - Houses `players`, `game_sessions`, `moves`, `player_stats`, `game_events` and the supporting migrations in `migrations/`
4. **Optional Kafka (Confluent Cloud)**
   - `internal/analytics` produces events; `cmd/analytics` consumes them to build dashboards or snapshots

## Key Source Files by Concern
| Concern | Primary File | Why It Matters |
| --- | --- | --- |
| WebSocket hub | [internal/websocket/hub.go](internal/websocket/hub.go) | Manages every connection, broadcasts, and room membership |
| WebSocket handler | [internal/websocket/handler.go](internal/websocket/handler.go) | Deserializes client messages and routes them to services |
| Matchmaking | [internal/matchmaking/service.go](internal/matchmaking/service.go) | Queue-based pairing, bot fallback after 10s |
| Game state | [internal/game/service.go](internal/game/service.go) | Session lifecycle, reconnections, custom rooms |
| Engine validation | [internal/game/engine.go](internal/game/engine.go) | Move validation, board checks, win detection |
| Player repo | [internal/database/repositories/player_repository.go](internal/database/repositories/player_repository.go) | Upserts players using Supabase profile data |
| Analytics producer | [internal/analytics/producer.go](internal/analytics/producer.go) | Writes events to Kafka while avoiding blocking gameplay |
| Frontend socket | [web/src/services/websocket.ts](web/src/services/websocket.ts) | Handles WS lifecycle, reconnect logic, and message dispatch |

## Running Locally vs Production
- **Local backend**: `make run-server` or `go run ./cmd/server` with `.env` pointing to local PostgreSQL and no Kafka credentials
- **Local WebSocket debug**: Run `web` via `npm install && npm run dev` (Vite) and connect to `http://localhost:8080/ws`
- **Migrations**: `go run ./cmd/migrate` runs `migrations/` scripts plus GORM AutoMigrate
- **Production**: Push to `main`; Render/GitHub Actions auto-deploy with Supabase + Kafka settings supplied via secrets

## Keeping It Fresh
- Link this file in `README.md` and add a checklist entry whenever you touch the hub, game service, or deployment
- Update the “Key Source Files” table if the project structure changes
- Mention new services (Redis, observability) here once added so this doc does not rot
