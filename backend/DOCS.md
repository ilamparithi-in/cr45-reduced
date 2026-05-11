# Backend Docs

## Overview

The backend is a Go HTTP API backed by PostgreSQL. It serves timetable data, accepts authenticated timetable changes, and runs SQL migrations automatically on startup.

## Location

- [backend/cmd/server/main.go](backend/cmd/server/main.go)
- [backend/internal/api/router.go](backend/internal/api/router.go)
- [backend/internal/store/postgres.go](backend/internal/store/postgres.go)
- [backend/internal/store/store.go](backend/internal/store/store.go)
- [backend/internal/auth/service.go](backend/internal/auth/service.go)
- [backend/internal/domain/types.go](backend/internal/domain/types.go)
- [backend/migrations/001_init.sql](backend/migrations/001_init.sql)
- [backend/migrations/002_seed.sql](backend/migrations/002_seed.sql)
- [backend/migrations/migrate.go](backend/migrations/migrate.go)

## Responsibilities

- Start the HTTP server
- Connect to PostgreSQL
- Run SQL migrations on startup
- Serve health and API routes
- Read timetable data from PostgreSQL
- Apply timetable overrides in PostgreSQL
- Delete slots and reindex later slot numbers
- Validate login credentials and issue signed tokens

## Runtime

- Default listen address: `:8080`
- Configurable with `HTTP_ADDR`
- Database DSN from `DATABASE_URL`
- Token signing secret from `APP_SECRET`

## Routes

- `GET /healthz`
  - returns `ok`
- `POST /api/auth/login`
  - accepts JSON username and password
  - returns a signed bearer token on success
- `GET /api/timetable?class_id=<id>`
  - returns the resolved timetable as JSON
- `POST /api/admin/override`
  - accepts JSON
  - requires `Content-Type: application/json`
  - requires `Authorization: Bearer <token>`
  - updates an existing slot when `slot_index` already exists
  - appends a new slot when `slot_index` is the next available index
- `DELETE /api/admin/slot?class_id=<id>&slot_index=<n>`
  - requires `Authorization: Bearer <token>`
  - deletes the selected slot
  - reindexes all later slots so numbering stays contiguous

## Response Headers

The backend sets these headers on JSON responses:

- `Content-Type: application/json`
- `Cache-Control: no-store`
- `X-Content-Type-Options: nosniff`

## Data Model

The database seed ships with a single default class ID:

- `107125`

The seeded timetable contains three initial slots.
Additional slots can be appended by submitting the next available slot index.
Deleting a slot shifts all later slot indices down by one in both the base data and override data.

## Authentication Model

- The login endpoint checks the submitted password against a stored hash from the seeded migration.
- Successful login returns a signed token.
- The frontend stores that token locally and sends it only on override requests.
- If the user is not logged in, the timetable remains visible but the override form is hidden.

## Local Development

### PostgreSQL

Create the local database before starting the backend:

```bash
createdb cr45_reduced
```

Alternative with `psql`:

```bash
psql -U postgres -c 'CREATE DATABASE cr45_reduced;'
```

If your local PostgreSQL credentials differ from the default connection string, export `DATABASE_URL` explicitly:

```bash
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/cr45_reduced?sslmode=disable'
```

### Run

```bash
cd backend
go mod tidy
go run ./cmd/server
```

If PostgreSQL is not running at the default DSN, start the backend with a custom database URL:

```bash
DATABASE_URL='postgres://postgres:postgres@localhost:5432/cr45_reduced?sslmode=disable' go run ./cmd/server
```

On successful startup, migrations create and seed the required tables. The seeded login for local testing is:

- username: `admin`
- password: `classrep123`

## What to Check When Debugging

- Migration failures or PostgreSQL connection errors in backend logs
- Response `Content-Type` when a request unexpectedly fails
- Login response token and `Authorization` header on override requests
- `POST /api/admin/override` request body and headers when updates return `415`
- `DELETE /api/admin/slot` request parameters and auth header when deletes fail