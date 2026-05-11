# Frontend Docs

## Overview

The frontend is a React application built with Vite. It displays the timetable, shows response headers for inspection, and exposes login-gated controls for updating and deleting slots.

## Location

- [frontend/src/App.jsx](frontend/src/App.jsx)
- [frontend/src/styles.css](frontend/src/styles.css)
- [frontend/vite.config.js](frontend/vite.config.js)

## Responsibilities

- Load the current timetable from the backend
- Allow login for privileged actions
- Submit override updates
- Delete slots after login
- Display API response headers for inspection
- Show failures clearly when responses are missing or malformed

## Runtime

- Vite dev server runs on `http://localhost:5173`
- The current Vite proxy forwards `/api` to `http://localhost:8080`

## Important Note

The frontend proxy target matches the backend default listen address.

## UI Behavior

- On load, the app requests `/api/timetable?class_id=107125`
- It displays timetable rows and response headers
- If not authenticated, it shows only the timetable and a login form
- After login, it reveals the override form
- The override form can update an existing slot or append the next slot
- Authenticated users can delete individual slots from the timetable table
- On submit, it sends a `POST /api/admin/override` request with JSON and bearer token
- On delete, it sends a `DELETE /api/admin/slot` request with bearer token
- If the backend response is not JSON or the request is blocked, the UI shows an error message

## Local Development

### Run

```bash
cd frontend
npm install
npm run dev
```

Frontend runs on `http://localhost:5173`.

## What to Check When Debugging

- Network tab for failed requests
- Console tab for browser-side header or CSP issues
- Whether the frontend proxy points to the correct backend port
- Login response token and whether authenticated requests include `Authorization: Bearer <token>`
- Whether the backend returns JSON when the UI expects JSON