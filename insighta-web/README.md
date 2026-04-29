# Insighta Web Portal

A Next.js web portal for the Insighta Labs+ Profile Intelligence Platform.

## Setup

```bash
npm install
```

Set the backend URL in `.env.local`:
```
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Running

```bash
npm run dev     # development
npm run build   # production build
npm start       # production server
```

## Pages

| Route | Description |
|-------|-------------|
| `/login` | GitHub OAuth login |
| `/dashboard` | Stats overview |
| `/profiles` | Profile list with filters and pagination |
| `/profiles/:id` | Profile detail view |
| `/search` | Natural language search |
| `/account` | Current user info and logout |

## Authentication

- Login redirects to the backend `/auth/github` endpoint
- Backend sets HTTP-only `access_token` and `refresh_token` cookies after OAuth
- All API calls use `withCredentials: true` so cookies are sent automatically
- On 401, the portal automatically calls `/auth/refresh` to get new tokens
- If refresh fails, the user is redirected to `/login`

## CSRF Protection

- The backend sets a `csrf_token` cookie (not HttpOnly) on API requests
- Mutating requests (POST, DELETE) read this cookie and send it as `X-CSRF-Token` header
- Bearer token requests (CLI) skip CSRF validation server-side
