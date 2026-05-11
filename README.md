# cr45-reduced

A full-stack app for you to play around with and dockerize

## Tech Stack

- Backend: Go
- Frontend: React + Vite
- Database: PostgreSQL

## Repository Layout

- `backend/` Go API server
- `frontend/` React app

## Documentation

Technical details are split by service:
- `backend/DOCS.md` for backend API, auth, migrations, and PostgreSQL notes
- `frontend/DOCS.md` for frontend runtime, proxy behavior, and UI flow

## Prerequisites

- Go 1.25+
- Node.js 20+
- npm 10+
- PostgreSQL 15+

## Run Locally

### PostgreSQL

Minimal local setup with `psql`:

```bash
createdb cr45_reduced
```

If `createdb` is not available, use:

```bash
psql -U postgres -c 'CREATE DATABASE cr45_reduced;'
```

If your local PostgreSQL user/password differs from the default DSN (Data Source Name), set `DATABASE_URL` before starting the backend.

Example:

```bash
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/cr45_reduced?sslmode=disable'
```

### Backend

```bash
cd backend
go mod tidy
go run ./cmd/server
```

Backend listens on `http://localhost:8080`.

Environment variables:
- `HTTP_ADDR` optional, default `:8080`
- `DATABASE_URL` optional, default `postgres://postgres:postgres@localhost:5432/cr45_reduced?sslmode=disable`
- `APP_SECRET` optional, default `cr45-reduced-dev-secret`

The backend runs SQL migrations automatically on startup.

After the backend starts successfully, the database will contain a seeded admin login created by migration:
- username: `admin`
- password: `classrep123`

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend runs on `http://localhost:5173`.

## Environment

Backend:
- `HTTP_ADDR` (optional, default `:8080`)
- `DATABASE_URL` (optional PostgreSQL DSN)
- `APP_SECRET` (optional token signing secret)

Frontend:
- `VITE_CLASS_ID` (optional, default `107125`)
