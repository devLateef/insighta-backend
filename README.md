# Insighta Labs+ — Backend

A secure, multi-interface Profile Intelligence platform built with Go.

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Clients                              │
│   CLI (insighta)    │   Web Portal    │   Direct API        │
└────────┬────────────┴────────┬────────┴────────┬────────────┘
         │                     │                  │
         └─────────────────────┼──────────────────┘
                               │  HTTPS
                    ┌──────────▼──────────┐
                    │   Gin HTTP Server   │
                    │   :8080             │
                    │                     │
                    │  Middleware stack:  │
                    │  • RequestLogger    │
                    │  • RateLimiter      │
                    │  • APIVersion       │
                    │  • JWTAuth          │
                    │  • CSRF             │
                    │  • RequireRole      │
                    └──────────┬──────────┘
                               │
                    ┌──────────▼──────────┐
                    │   PostgreSQL 15     │
                    │   • users           │
                    │   • refresh_tokens  │
                    │   • profiles        │
                    └─────────────────────┘
```

### Repositories

| Repo | Description |
|------|-------------|
| `insighta-backend` | This repo — Go API server |
| `insighta-cli` | Globally installable CLI tool |
| `insighta-web` | Web portal (Next.js) |

---

## Authentication Flow

### Web Flow (browser)

```
Browser → GET /auth/github
       ← 302 redirect to GitHub OAuth
GitHub → GET /auth/github/callback?code=...&state=...
       ← Sets HTTP-only cookies: access_token, refresh_token
       ← 302 redirect to /dashboard
```

### CLI Flow (PKCE)

```
CLI generates:
  state          = random 32-byte base64url
  code_verifier  = random 32-byte base64url
  code_challenge = SHA-256(code_verifier) base64url

CLI → GET /auth/github?state=...&code_challenge=...
    ← 302 redirect to GitHub OAuth

GitHub → local callback server at http://localhost:<port>/callback
       captures code + state

CLI validates state, then:
CLI → GET /auth/github/callback?code=...&code_verifier=...
    ← JSON: { access_token, refresh_token, user }

CLI stores tokens at ~/.insighta/credentials.json
```

### Token Lifecycle

| Token | Expiry | Storage |
|-------|--------|---------|
| Access token | 3 minutes | Authorization header / HTTP-only cookie |
| Refresh token | 5 minutes | Server DB (hashed) + HTTP-only cookie / credentials file |

- Refresh tokens are **rotated on every use** — the old token is invalidated immediately
- Refresh tokens are stored as SHA-256 hashes in the database (never plaintext)
- On refresh, a fresh user record is fetched from DB so role changes take effect immediately

---

## API Endpoints

All `/api/*` endpoints require:
- `X-API-Version: 1` header
- Valid JWT access token (Bearer or cookie)

### Auth

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/auth/github` | — | Redirect to GitHub OAuth |
| GET | `/auth/github/callback` | — | OAuth callback, issues tokens |
| POST | `/auth/refresh` | — | Rotate refresh token |
| POST | `/auth/logout` | — | Invalidate refresh token |
| GET | `/auth/me` | JWT | Current user info |

### Profiles

| Method | Path | Role | Description |
|--------|------|------|-------------|
| GET | `/api/profiles` | any | List profiles (filter/sort/paginate) |
| GET | `/api/profiles/search?q=` | any | Natural language search |
| GET | `/api/profiles/:id` | any | Get single profile |
| GET | `/api/profiles/export?format=csv` | any | Export CSV |
| POST | `/api/profiles` | admin | Create profile |
| DELETE | `/api/profiles/:id` | admin | Delete profile |

### Query Parameters (list & search)

| Param | Example | Description |
|-------|---------|-------------|
| `gender` | `male` | Filter by gender |
| `country_id` | `NG` | Filter by ISO 2-letter country code |
| `age_group` | `adult` | Filter by age group (child/teenager/adult/senior) |
| `min_age` | `25` | Minimum age (inclusive) |
| `max_age` | `40` | Maximum age (inclusive) |
| `min_gender_probability` | `0.9` | Minimum gender confidence score (0–1) |
| `min_country_probability` | `0.8` | Minimum country confidence score (0–1) |
| `sort_by` | `age` | Sort field: `age`, `created_at`, `gender_probability`, `country_probability` |
| `order` | `asc` | Sort direction: `asc` or `desc` |
| `page` | `2` | Page number (default: 1) |
| `limit` | `20` | Page size (default: 10, max: 50) |

### Pagination Response Shape

```json
{
  "status": "success",
  "page": 1,
  "limit": 10,
  "total": 2026,
  "total_pages": 203,
  "links": {
    "self": "/api/profiles?page=1&limit=10",
    "next": "/api/profiles?page=2&limit=10",
    "prev": null
  },
  "data": [...]
}
```

---

## Role Enforcement

Two roles exist: `admin` and `analyst`.

| Action | admin | analyst |
|--------|-------|---------|
| List / search / get profiles | ✅ | ✅ |
| Export CSV | ✅ | ✅ |
| Create profile | ✅ | ❌ |
| Delete profile | ✅ | ❌ |

- Default role for new users: `analyst`
- Role is always read from the **database** on every request (not from the JWT claim), so role changes take effect immediately without re-login
- RBAC is enforced via `middleware.RequireRole()` applied at the route level

---

## Natural Language Parsing

`GET /api/profiles/search?q=young males from nigeria`

### How it works

The parser is **fully rule-based** — no AI, no LLMs. It tokenizes the query into lowercase words, then applies pattern matching in this order:

**1. Gender detection**

| Keywords | Maps to |
|----------|---------|
| `male`, `males`, `men`, `man`, `boy`, `boys` | `gender=male` |
| `female`, `females`, `women`, `woman`, `girl`, `girls` | `gender=female` |

**2. Age group detection**

| Keywords | Maps to |
|----------|---------|
| `young` | `min_age=16 + max_age=24` (parsing only — not a stored age_group) |
| `teenager`, `teenagers`, `teen`, `teens` | `age_group=teenager` |
| `adult`, `adults` | `age_group=adult` |
| `child`, `children`, `kid`, `kids` | `age_group=child` |
| `senior`, `seniors`, `elderly`, `old` | `age_group=senior` |

**3. Age modifier detection**

| Pattern | Maps to |
|---------|---------|
| `above N` / `over N` | `min_age=N` |
| `below N` / `under N` | `max_age=N` |
| `older than N` | `min_age=N` |
| `younger than N` | `max_age=N` |

**4. Country detection**

- `from <country>` — matches country names and common aliases (e.g. `from nigeria` → `country_id=NG`, `from usa` → `country_id=US`)
- Bare demonyms anywhere in the query (e.g. `nigerian`, `kenyan`, `ghanaian`)
- Supports 60+ countries including all major African, European, Asian, and American nations

**Example mappings**

| Query | Parsed filters |
|-------|---------------|
| `young males from nigeria` | `gender=male, min_age=16, max_age=24, country_id=NG` |
| `females above 30` | `gender=female, min_age=30` |
| `people from angola` | `country_id=AO` |
| `adult males from kenya` | `gender=male, age_group=adult, country_id=KE` |
| `male and female teenagers above 17` | `age_group=teenager, min_age=17` |
| `nigerian males` | `gender=male, country_id=NG` |
| `seniors from ghana` | `age_group=senior, country_id=GH` |
| `men over 25` | `gender=male, min_age=25` |
| `women under 40` | `gender=female, max_age=40` |

**Uninterpretable queries** return:
```json
{ "status": "error", "message": "Unable to interpret query" }
```

A query is uninterpretable if none of the above patterns match any token.

### Limitations

- **No compound age ranges from single words**: `"middle-aged"` is not supported
- **No negation**: `"not from nigeria"` is not handled
- **No OR logic for countries**: `"from nigeria or ghana"` only picks up the first match
- **Ambiguous demonyms**: `"thai"` maps to Thailand but `"chinese"` could be ambiguous in some contexts
- **No fuzzy matching**: typos like `"nigerria"` will not match
- **"young" is a parsing alias only**: it maps to `min_age=16, max_age=24` and does not correspond to a stored `age_group` value
- **Multi-word country names after "from"**: `"from south africa"` works, but `"south african people"` requires the demonym form `"south african"` which is not currently in the adjective map (use `"from south africa"` instead)

---

## Rate Limiting

| Scope | Limit | Key |
|-------|-------|-----|
| `/auth/*` | 10 req/min | Per IP |
| `/api/*` | 60 req/min | Per user ID (falls back to IP) |

Returns `429 Too Many Requests` when exceeded.

---

## Running Locally

### Prerequisites
- Go 1.22+
- Docker & Docker Compose
- GitHub OAuth App ([create one here](https://github.com/settings/developers))

### Setup

```bash
# 1. Clone the repo
git clone https://github.com/your-org/insighta-backend
cd insighta-backend

# 2. Configure environment
cp .env.example .env
# Edit .env — set GITHUB_CLIENT_ID, GITHUB_SECRET, JWT_SECRET

# 3. Start database + backend
docker compose up --build

# 4. Or run locally (requires Postgres running)
go run ./cmd/server
```

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `PORT` | No | Server port (default: 8080) |
| `JWT_SECRET` | Yes | Secret for signing JWTs |
| `GITHUB_CLIENT_ID` | Yes | GitHub OAuth App client ID |
| `GITHUB_SECRET` | Yes | GitHub OAuth App client secret |
| `BASE_URL` | Yes | Public base URL (used for OAuth redirect) |
| `DATABASE_URL` | Yes | PostgreSQL connection string |

### GitHub OAuth App Setup

1. Go to **GitHub → Settings → Developer settings → OAuth Apps → New OAuth App**
2. Set **Authorization callback URL** to `{BASE_URL}/auth/github/callback`
3. Copy **Client ID** and generate a **Client Secret**
4. Paste both into `.env`

---

## Running Tests

```bash
go test ./... -v
```

Tests cover JWT generation/validation, PKCE utilities, middleware (API versioning, JWT auth, RBAC), and the natural language parser.

---

## Seeding the Database

Download the 2026 profiles JSON file and run:

```bash
# First time
go run ./cmd/seed --file=profiles.json

# Re-run safely (duplicates are skipped via ON CONFLICT DO NOTHING)
go run ./cmd/seed --file=profiles.json

# Reset and re-seed
go run ./cmd/seed --file=profiles.json --reset
```

The seed command expects a JSON array of profile objects with fields matching the `profiles` table schema.

---

## Deployment

The backend is deployed at: `https://your-backend-url.com`

The web portal is deployed at: `https://your-portal-url.com`

---

## Project Structure

```
insighta-backend/
├── cmd/server/          # Entry point
├── configs/             # Config loading
├── internal/
│   ├── handlers/        # HTTP handlers (auth, profiles, export, external APIs)
│   ├── middleware/       # JWT auth, RBAC, rate limiting, CSRF, logging, versioning
│   ├── models/          # Data models
│   ├── storage/         # Database layer
│   └── utils/           # PKCE helpers
├── migrations/          # SQL migrations
├── pkg/jwt/             # JWT package
├── .github/workflows/   # CI/CD
├── Dockerfile
├── docker-compose.yml
└── README.md
```
