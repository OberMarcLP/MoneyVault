# MoneyVault

A self-hosted, encrypted personal finance platform for budgeting, investments, and crypto tracking.

## Tech Stack

- **Backend:** Go (Gin), PostgreSQL 16, AES-256-GCM encryption at rest
- **Frontend:** React 19 (Vite), Tailwind CSS 4, TanStack Query, Zustand
- **Auth:** Argon2id password hashing, JWT access/refresh tokens, TOTP 2FA
- **Deployment:** Docker Compose (Caddy + Go API + PostgreSQL)

## Quick Start (Development)

### Prerequisites

- Go 1.26+
- Node.js 20+
- Docker & Docker Compose
- PostgreSQL 16 (or use Docker)

### 1. Start the database

```bash
docker compose -f docker-compose.dev.yml up db -d
```

### 2. Run the backend

```bash
cd backend
cp ../.env.example .env  # Edit with your settings
go run ./cmd/server
```

### 3. Run the frontend

```bash
cd frontend
npm install
npm run dev
```

The app will be available at `http://localhost:5173`.

## Production Deployment

```bash
cp .env.example .env
# Edit .env with strong passwords and secrets

# Build the frontend
cd frontend && npm run build && cd ..

# Start everything
docker compose up -d
```

### Running Behind a Reverse Proxy

If you have an existing reverse proxy (nginx, Traefik, Caddy, etc.) that handles TLS termination:

1. **Set environment variables** in `.env`:

```env
ALLOWED_ORIGIN=https://your-domain.com
SITE_ADDRESS=:80
AUTO_HTTPS=off
TRUSTED_PROXIES=172.18.0.0/16
# Optional: change host port if 80 is taken
HTTP_PORT=8180
```

2. **Start the stack:**

```bash
docker compose up -d
```

3. **Configure your reverse proxy** to forward to `http://localhost:${HTTP_PORT:-80}`.

**nginx example:**

```nginx
location / {
    proxy_pass http://127.0.0.1:8180;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

**Find your Docker bridge CIDR** (for `TRUSTED_PROXIES`):

```bash
docker network inspect moneyvault_default | grep Subnet
```

## Project Structure

```
moneyvault/
├── backend/               # Go API server
│   ├── cmd/server/        # Entry point
│   ├── internal/
│   │   ├── config/        # Environment configuration
│   │   ├── encryption/    # AES-256-GCM encryption
│   │   ├── handlers/      # HTTP handlers
│   │   ├── middleware/     # Auth, CORS, rate limiting
│   │   ├── models/        # Data structures
│   │   ├── repositories/  # Database queries
│   │   └── services/      # Business logic
│   └── migrations/        # SQL migrations
├── frontend/              # React SPA
│   └── src/
│       ├── api/           # API client & TanStack Query hooks
│       ├── components/    # UI components, layout, forms
│       ├── pages/         # Route-level pages
│       ├── stores/        # Zustand state stores
│       └── types/         # TypeScript types
├── docker-compose.yml      # Production deployment (GHCR images)
├── docker-compose.dev.yml  # Development (local build)
└── docker-compose.test.yml # Test database (port 5433)
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Create account |
| POST | `/api/v1/auth/login` | Sign in |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/logout` | Sign out |
| GET | `/api/v1/auth/me` | Get current user |
| PUT | `/api/v1/auth/preferences` | Update preferences |
| POST | `/api/v1/auth/totp/setup` | Set up 2FA |
| POST | `/api/v1/auth/totp/verify` | Enable 2FA |
| DELETE | `/api/v1/auth/totp` | Disable 2FA |
| CRUD | `/api/v1/accounts` | Financial accounts |
| CRUD | `/api/v1/transactions` | Transactions |
| CRUD | `/api/v1/categories` | Categories |

## Security

- All sensitive fields (balances, amounts, descriptions) are encrypted at rest using AES-256-GCM
- Per-user Data Encryption Keys (DEK), encrypted with Argon2id-derived keys
- JWT access tokens (15 min) + httpOnly refresh token cookies (7 days)
- TOTP two-factor authentication
- Rate limiting, CORS protection, parameterized queries
