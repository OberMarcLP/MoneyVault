# MoneyVault — Project Plan

> A self-hosted, encrypted personal finance platform for budgeting, investments, and crypto tracking.

---

## 1. Project Overview

**MoneyVault** is a self-hosted, open-source personal finance platform that gives users full control over their financial data. It combines budgeting, stock/ETF portfolio tracking, crypto (including DeFi & staking), and bank import capabilities — all behind end-to-end encryption with a modern, themeable UI.

### Core Principles

- **Privacy-first:** Self-hosted, E2E encrypted, zero third-party data sharing
- **All-in-one:** Budget, invest, and track crypto from a single dashboard
- **Modern UX:** Clean interface with dark/light mode, responsive + installable PWA
- **Multi-user:** Role-based access for families, partners, or teams
- **Dockerized:** One-command deployment via Docker Compose

---

## 2. Tech Stack

| Layer            | Technology                          | Rationale                                                |
| ---------------- | ----------------------------------- | -------------------------------------------------------- |
| **Frontend**     | React 18+ (Vite)                    | Fast builds, modern DX, huge ecosystem                   |
| **Styling**      | Tailwind CSS                        | Utility-first, easy dark/light theming                   |
| **State Mgmt**   | Zustand or TanStack Query           | Lightweight, great for async server state                |
| **Charts**       | Recharts or Tremor                  | React-native charting, clean financial visuals            |
| **PWA**          | Workbox (via Vite PWA plugin)       | Service worker caching, installability, push notifications|
| **Backend**      | Go (Gin or Echo framework)          | High performance, low memory, great concurrency          |
| **Database**     | PostgreSQL 16+                      | ACID compliance, JSON support, excellent for financial data|
| **Encryption**   | AES-256-GCM (at rest) + NaCl/libsodium (E2E) | Industry standard, proven cryptography         |
| **Auth**         | Custom (bcrypt/argon2 + JWT + WebAuthn) | Email/password + TOTP 2FA + Passkeys                 |
| **Containerization** | Docker + Docker Compose         | Single-command self-hosted deployment                    |
| **Reverse Proxy**| Caddy or Traefik                    | Auto HTTPS, easy config                                  |
| **Task Queue**   | Go routines + Redis (optional)      | Background jobs for price fetching, notifications        |

---

## 3. Architecture

```
┌─────────────────────────────────────────────────────┐
│                   Docker Compose                     │
│                                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │
│  │   Caddy       │  │   Frontend   │  │   Backend  │ │
│  │   (Reverse    │──│   React SPA  │  │   Go API   │ │
│  │    Proxy)     │  │   (Nginx)    │  │   Server   │ │
│  └──────────────┘  └──────────────┘  └─────┬──────┘ │
│                                            │        │
│                          ┌─────────────────┼────┐   │
│                          │                 │    │   │
│                    ┌─────▼─────┐   ┌──────▼──┐ │   │
│                    │ PostgreSQL │   │  Redis   │ │   │
│                    │ (encrypted │   │ (cache/  │ │   │
│                    │  at rest)  │   │  queue)  │ │   │
│                    └───────────┘   └─────────┘ │   │
│                                                 │   │
│                    ┌────────────────────────────┘   │
│                    │  Background Workers             │
│                    │  - Price fetcher (stocks/crypto) │
│                    │  - Notification dispatcher       │
│                    │  - Import processor              │
│                    └─────────────────────────────────┘
└─────────────────────────────────────────────────────┘
```

### API Design

- RESTful JSON API with versioning (`/api/v1/...`)
- JWT access tokens (short-lived) + refresh tokens (long-lived, httpOnly cookies)
- Rate limiting per user/role
- OpenAPI/Swagger documentation auto-generated

---

## 4. Database Schema (High-Level)

### Core Tables

```
users
├── id (UUID)
├── email (encrypted)
├── password_hash (argon2id)
├── role (admin | user | viewer)
├── totp_secret (encrypted)
├── passkey_credentials (JSONB)
├── encryption_key_salt
├── created_at / updated_at
└── preferences (JSONB: theme, currency, locale)

accounts
├── id (UUID)
├── user_id (FK)
├── name (encrypted)
├── type (checking | savings | credit | investment | crypto_wallet | defi)
├── currency
├── balance (encrypted)
├── institution
└── is_active

transactions
├── id (UUID)
├── account_id (FK)
├── user_id (FK)
├── type (income | expense | transfer)
├── amount (encrypted)
├── currency
├── category_id (FK)
├── description (encrypted)
├── date
├── tags (JSONB)
├── import_source (manual | csv | api)
└── created_at

categories
├── id (UUID)
├── user_id (FK)
├── name
├── type (income | expense)
├── icon
├── color
└── parent_id (FK, for subcategories)

budgets
├── id (UUID)
├── user_id (FK)
├── category_id (FK)
├── amount
├── period (monthly | weekly | yearly)
├── start_date / end_date
└── rollover (boolean)

holdings (stocks/ETFs/crypto)
├── id (UUID)
├── user_id (FK)
├── account_id (FK)
├── asset_type (stock | etf | crypto | defi_position)
├── symbol / token_address
├── quantity
├── avg_cost_basis
├── acquired_at
└── metadata (JSONB: staking info, pool data, etc.)

price_history
├── id
├── symbol
├── asset_type
├── price
├── currency
├── fetched_at
└── source (yahoo | coingecko)

notifications
├── id (UUID)
├── user_id (FK)
├── type (budget_alert | price_alert | milestone)
├── title / message
├── is_read
├── delivered_via (push | in_app)
└── created_at

alert_rules
├── id (UUID)
├── user_id (FK)
├── type (budget_overspend | price_drop | price_rise | net_worth_milestone)
├── condition (JSONB)
├── is_active
└── created_at
```

---

## 5. Encryption Strategy

### End-to-End Encryption (E2E)

```
User Password
     │
     ▼
  PBKDF2/Argon2 ──► Master Key (derived, never stored on server)
     │
     ▼
  Encrypts/Decrypts ──► Data Encryption Key (DEK)
     │
     ▼
  DEK encrypts ──► Sensitive fields (balances, amounts, descriptions, etc.)
```

- **Client-side encryption:** Sensitive data is encrypted in the browser before being sent to the server
- **Server never sees plaintext** financial data
- **Key derivation:** User's password derives a master key via Argon2id
- **Data Encryption Key (DEK):** Random AES-256-GCM key, encrypted with the master key and stored server-side
- **Key rotation:** Support for periodic DEK rotation without re-encrypting all data immediately

### Encryption at Rest

- PostgreSQL with `pgcrypto` extension for column-level encryption as a fallback
- Full disk encryption on Docker volumes (recommended in deployment docs)
- Redis configured with encrypted connections (TLS)

### What Gets Encrypted

| Field                     | E2E Encrypted | At Rest |
| ------------------------- | ------------- | ------- |
| Account balances          | ✅            | ✅      |
| Transaction amounts       | ✅            | ✅      |
| Transaction descriptions  | ✅            | ✅      |
| User email                | ✅            | ✅      |
| Account names             | ✅            | ✅      |
| Holdings quantities       | ✅            | ✅      |
| Category names            | ❌            | ✅      |
| Price history             | ❌            | ❌      |
| Alert rules               | ❌            | ✅      |

---

## 6. Feature Breakdown

### Phase 1 — Foundation (Weeks 1–4)

> Goal: Core app running in Docker with auth, accounts, and manual transactions.

- [ ] **Project scaffolding**
  - Go backend with Gin/Echo, project structure (handlers, services, repositories, middleware)
  - React frontend with Vite, Tailwind, routing (React Router)
  - Docker Compose (Go API + React Nginx + PostgreSQL + Redis)
  - Database migrations (golang-migrate or goose)

- [ ] **Authentication system**
  - Email/password registration & login (Argon2id hashing)
  - JWT access + refresh token flow
  - TOTP 2FA setup & verification
  - WebAuthn/Passkey registration & login
  - Session management & device tracking

- [ ] **User management & roles**
  - Admin panel for user creation/invitation
  - Role-based access control (admin, user, viewer)
  - Viewer role: read-only access to shared dashboards
  - User preferences (default currency, locale, theme)

- [ ] **Dark/Light theme**
  - Tailwind dark mode with system preference detection
  - Theme toggle with persistence
  - CSS custom properties for accent colors

- [ ] **Account management**
  - CRUD for financial accounts (checking, savings, credit, investment, crypto)
  - Account grouping by institution
  - Multi-currency support

- [ ] **Manual transactions**
  - Add/edit/delete income & expense transactions
  - Category assignment with icons
  - Tags and notes
  - Transfer between accounts
  - Recurring transaction templates

- [ ] **E2E encryption layer**
  - Client-side encryption/decryption library
  - Key derivation from password
  - Encrypted field storage in PostgreSQL

### Phase 2 — Budgeting & Imports (Weeks 5–8)

> Goal: Full budgeting system with file imports.

- [ ] **Budget system**
  - Create budgets per category (monthly/weekly/yearly)
  - Budget rollover (carry unused budget to next period)
  - Budget vs actual spending dashboard
  - Progress bars and visual indicators
  - Overspend warnings

- [ ] **Category management**
  - Default category templates (Food, Transport, Housing, etc.)
  - Custom categories with icons and colors
  - Subcategory support (Food → Groceries, Restaurants)
  - Auto-categorization rules (keyword matching)

- [ ] **File import engine**
  - CSV parser with column mapping UI (drag & drop columns)
  - OFX/QIF file parser
  - Duplicate detection (fuzzy matching on date + amount + description)
  - Import history & undo
  - Saved import profiles per bank

- [ ] **Dashboard v1**
  - Net worth summary card
  - Monthly income vs expenses
  - Recent transactions list
  - Budget overview cards
  - Account balances at a glance

### Phase 3 — Investments & Stocks (Weeks 9–12)

> Goal: Track stock/ETF portfolios with performance analytics.

- [ ] **Stock/ETF portfolio tracker**
  - Add holdings manually (symbol, quantity, cost basis, date)
  - Portfolio overview with current values
  - Holdings breakdown (pie chart by asset/sector)
  - Support multiple investment accounts

- [ ] **Price data integration**
  - Yahoo Finance API integration (free tier)
  - Background worker: fetch prices on schedule (every 15–30 min during market hours)
  - Price history storage & caching
  - Fallback handling when API is unavailable

- [ ] **Investment analytics**
  - Total return (absolute + percentage)
  - Unrealized vs realized gains
  - Performance over time (line charts: 1W, 1M, 3M, 1Y, ALL)
  - Dividend tracking
  - Cost basis methods (FIFO, LIFO, average)

- [ ] **Tax reporting: Capital gains**
  - Realized gains/losses report by tax year
  - Short-term vs long-term classification
  - Export to CSV for tax filing
  - Wash sale detection (future enhancement)

### Phase 4 — Crypto (Weeks 13–16)

> Goal: Full-depth crypto tracking with DeFi and staking.

- [ ] **Crypto portfolio tracker**
  - Manual entry of holdings
  - Exchange API connections (read-only API keys)
    - Binance
    - Coinbase
    - Kraken (extensible to more)
  - Auto-sync balances and transaction history

- [ ] **Price data: CoinGecko**
  - CoinGecko free API integration
  - Support for thousands of tokens
  - Historical price data for cost basis calculation
  - Rate limiting and caching

- [ ] **DeFi tracking**
  - Liquidity pool positions (token pairs, pool value)
  - Yield farming / staking rewards tracking
  - Manual entry with metadata fields
  - DeFi protocol tagging (Uniswap, Aave, Lido, etc.)

- [ ] **On-chain history**
  - Wallet address tracking (EVM chains: Ethereum, Polygon, BSC, Arbitrum)
  - Transaction history via public APIs (Etherscan, etc.)
  - Token transfer tracking
  - Gas fee tracking

- [ ] **Crypto tax reporting**
  - Cost basis tracking across exchanges and wallets
  - Capital gains calculation for crypto trades
  - DeFi income classification (staking rewards = income)
  - Export for tax purposes

### Phase 5 — Analytics & Reports (Weeks 17–20)

> Goal: Comprehensive financial reports and insights.

- [ ] **Net worth over time**
  - Daily/weekly/monthly net worth snapshots
  - Net worth chart with breakdown by account type
  - Milestone tracking (e.g., "reached $100k")

- [ ] **Spending breakdowns**
  - By category, subcategory, merchant
  - Monthly/quarterly/yearly comparisons
  - Spending trends and averages
  - Top expenses ranking

- [ ] **Investment performance**
  - Portfolio performance vs benchmarks (S&P 500, BTC)
  - Asset allocation analysis
  - Risk/diversification indicators
  - Sector/geography breakdown

- [ ] **Budget vs actual**
  - Monthly comparison charts
  - Category-level drill-down
  - Historical budget adherence tracking
  - Insights: "You overspent on dining 3 of the last 6 months"

- [ ] **Cash flow forecasting**
  - Projected income vs expenses based on recurring transactions
  - "What if" scenarios (e.g., add a new expense, see impact)
  - Runway calculation: "At this burn rate, savings last X months"
  - Upcoming bills calendar view

### Phase 6 — Notifications & PWA (Weeks 21–24)

> Goal: Push notifications, alerts, and installable PWA.

- [ ] **In-app notification center**
  - Bell icon with unread count
  - Notification list (filterable by type)
  - Mark as read / dismiss / clear all
  - Notification preferences per type

- [ ] **Push notifications (Web Push API)**
  - Browser push notification support
  - VAPID key setup for web push
  - PWA push notification handling via service worker
  - Notification types:
    - Budget overspend alerts
    - Price alerts (stock/crypto above/below threshold)
    - Portfolio milestone alerts
    - Weekly/monthly summary digests
    - Import completion notifications

- [ ] **Alert rules engine**
  - User-configurable alert rules
  - Conditions: budget %, price threshold, net worth milestone
  - Evaluation frequency (real-time, hourly, daily)
  - Snooze / disable rules

- [ ] **PWA enhancements**
  - Service worker with Workbox (cache strategies)
  - Web app manifest (icons, splash screen, theme color)
  - Offline support for cached data viewing
  - Install prompt handling
  - Background sync for queued transactions

---

## 7. Project Structure

```
moneyvault/
├── docker-compose.yml
├── docker-compose.dev.yml
├── .env.example
├── README.md
│
├── backend/
│   ├── Dockerfile
│   ├── go.mod / go.sum
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── config/          # App configuration
│   │   ├── middleware/       # Auth, CORS, rate limiting, logging
│   │   ├── handlers/        # HTTP handlers (controllers)
│   │   │   ├── auth.go
│   │   │   ├── accounts.go
│   │   │   ├── transactions.go
│   │   │   ├── budgets.go
│   │   │   ├── holdings.go
│   │   │   ├── reports.go
│   │   │   └── notifications.go
│   │   ├── services/        # Business logic
│   │   ├── repositories/    # Database queries
│   │   ├── models/          # Data structures
│   │   ├── crypto/          # Encryption utilities
│   │   ├── importers/       # CSV/OFX/QIF parsers
│   │   ├── integrations/    # Yahoo Finance, CoinGecko, exchange APIs
│   │   └── workers/         # Background jobs
│   ├── migrations/          # SQL migration files
│   └── tests/
│
├── frontend/
│   ├── Dockerfile
│   ├── nginx.conf
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   ├── public/
│   │   ├── manifest.json
│   │   ├── sw.js
│   │   └── icons/
│   └── src/
│       ├── main.tsx
│       ├── App.tsx
│       ├── api/             # API client & hooks
│       ├── components/      # Reusable UI components
│       │   ├── ui/          # Buttons, inputs, modals, cards
│       │   ├── charts/      # Chart components
│       │   ├── layout/      # Sidebar, header, theme toggle
│       │   └── forms/       # Transaction, budget, holding forms
│       ├── pages/           # Route-level pages
│       │   ├── Dashboard/
│       │   ├── Accounts/
│       │   ├── Transactions/
│       │   ├── Budgets/
│       │   ├── Investments/
│       │   ├── Crypto/
│       │   ├── Reports/
│       │   ├── Settings/
│       │   └── Admin/
│       ├── hooks/           # Custom React hooks
│       ├── stores/          # Zustand stores
│       ├── utils/           # Helpers, formatters, encryption
│       ├── types/           # TypeScript types
│       └── styles/          # Global styles, Tailwind overrides
│
└── docs/
    ├── API.md
    ├── DEPLOYMENT.md
    ├── ENCRYPTION.md
    └── CONTRIBUTING.md
```

---

## 8. Docker Compose (Production)

```yaml
version: "3.9"

services:
  caddy:
    image: caddy:2-alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
    depends_on:
      - frontend
      - backend

  frontend:
    build: ./frontend
    expose:
      - "80"
    depends_on:
      - backend

  backend:
    build: ./backend
    expose:
      - "8080"
    environment:
      - DATABASE_URL=postgres://moneyvault:${DB_PASSWORD}@db:5432/moneyvault?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
      - ENCRYPTION_KEY=${ENCRYPTION_KEY}
      - VAPID_PUBLIC_KEY=${VAPID_PUBLIC_KEY}
      - VAPID_PRIVATE_KEY=${VAPID_PRIVATE_KEY}
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_started

  db:
    image: postgres:16-alpine
    volumes:
      - pg_data:/var/lib/postgresql/data
    environment:
      - POSTGRES_DB=moneyvault
      - POSTGRES_USER=moneyvault
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U moneyvault"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data

volumes:
  pg_data:
  redis_data:
  caddy_data:
```

---

## 9. Security Checklist

- [ ] Argon2id password hashing (memory: 64MB, iterations: 3, parallelism: 4)
- [ ] JWT tokens: 15-min access token, 7-day refresh token (httpOnly, secure, sameSite)
- [ ] TOTP 2FA with backup codes
- [ ] WebAuthn/Passkeys via `go-webauthn` library
- [ ] CSRF protection on all state-changing endpoints
- [ ] Rate limiting: login (5/min), API (100/min per user)
- [ ] Input validation & SQL injection prevention (parameterized queries)
- [ ] CORS restricted to frontend origin only
- [ ] Helmet-style security headers (CSP, HSTS, X-Frame-Options)
- [ ] API key encryption for exchange connections (encrypted at rest, never logged)
- [ ] Audit log for sensitive actions (login, export, role change)
- [ ] Automatic session invalidation on password change
- [ ] Docker: non-root containers, read-only filesystem where possible
- [ ] Secrets via environment variables (never committed to repo)

---

## 10. Free API Integrations

### Yahoo Finance (Stocks/ETFs)

- **Endpoint:** Unofficial Yahoo Finance API (via `yahoo-finance2` or direct HTTP)
- **Data:** Real-time quotes, historical prices, dividends, splits
- **Rate limit:** ~2000 requests/hour (unofficial, be conservative)
- **Strategy:** Fetch during market hours every 15–30 min, cache aggressively

### CoinGecko (Crypto)

- **Endpoint:** `api.coingecko.com/api/v3/`
- **Data:** Current prices, historical data, token metadata, DeFi stats
- **Rate limit:** 10–30 calls/min (free tier)
- **Strategy:** Batch requests (up to 250 coins per call), cache for 60s

### Exchange APIs (Crypto)

| Exchange  | Library / SDK     | Capabilities (read-only)                      |
| --------- | ----------------- | --------------------------------------------- |
| Binance   | Go binance client | Balances, trade history, staking info          |
| Coinbase  | Coinbase API v2   | Balances, transaction history                  |
| Kraken    | Kraken REST API   | Balances, trade history, staking               |

### Blockchain Explorers (On-chain)

| Chain     | API                     | Use                                    |
| --------- | ----------------------- | -------------------------------------- |
| Ethereum  | Etherscan API (free)    | Tx history, token transfers, gas fees  |
| Polygon   | Polygonscan API (free)  | Same as above                          |
| BSC       | BscScan API (free)      | Same as above                          |
| Arbitrum  | Arbiscan API (free)     | Same as above                          |

---

## 11. UI/UX Guidelines

### Design System

- **Typography:** Inter or Geist font family
- **Color palette:**
  - Light mode: White backgrounds, slate-gray text, blue accents
  - Dark mode: Slate-900 backgrounds, slate-100 text, blue/emerald accents
  - Positive values: Green (`#10B981`)
  - Negative values: Red (`#EF4444`)
- **Components:** shadcn/ui-inspired component library (built with Tailwind)
- **Spacing:** 4px grid system
- **Border radius:** Rounded-lg (8px) for cards, rounded-md (6px) for inputs

### Key Screens

1. **Dashboard** — Net worth card, monthly overview, recent transactions, budget rings, portfolio snapshot
2. **Accounts** — List of all accounts grouped by type, quick balance view
3. **Transactions** — Searchable/filterable table, quick-add button, import button
4. **Budgets** — Category cards with progress bars, monthly timeline
5. **Investments** — Portfolio value chart, holdings table, performance metrics
6. **Crypto** — Wallet overview, exchange balances, DeFi positions, staking rewards
7. **Reports** — Tab-based: Net Worth | Spending | Investments | Budget | Cash Flow | Tax
8. **Settings** — Profile, security (2FA, passkeys), preferences, connected exchanges, import profiles
9. **Admin** — User management, role assignment, system health, audit log

---

## 12. Development Timeline

| Phase | Weeks   | Milestone                                      |
| ----- | ------- | ---------------------------------------------- |
| 1     | 1 – 4   | Auth, accounts, manual transactions, encryption, Docker |
| 2     | 5 – 8   | Budgeting system, file imports, dashboard v1    |
| 3     | 9 – 12  | Stock/ETF tracking, Yahoo Finance, investment analytics |
| 4     | 13 – 16 | Crypto full depth, exchange APIs, DeFi, on-chain |
| 5     | 17 – 20 | Analytics suite, all reports, tax reporting      |
| 6     | 21 – 24 | Notifications, push alerts, PWA finalization     |
| —     | 25 – 26 | Testing, bug fixes, documentation, v1.0 release  |

**Estimated total: ~6 months** (at a steady pace with 1-2 devs)

---

## 13. Testing Strategy

- **Backend:** Go `testing` package + `testify` for unit & integration tests
- **Frontend:** Vitest + React Testing Library for component tests
- **E2E:** Playwright for critical user flows (login, add transaction, view reports)
- **API:** Bruno or Hurl for API endpoint testing
- **Security:** OWASP ZAP scan, manual penetration testing before v1.0
- **CI/CD:** GitHub Actions — lint, test, build, Docker image push

---

## 14. Future Enhancements (Post v1.0)

- [ ] AI-powered auto-categorization (local LLM or rule-based ML)
- [ ] Receipt scanning (OCR via Tesseract)
- [ ] Multi-currency with live exchange rates
- [ ] Shared budgets between users (partner mode)
- [ ] Plaid/GoCardless integration for live bank syncing
- [ ] Mobile app (React Native, sharing the API)
- [ ] Data export (full encrypted backup + restore)
- [ ] Plugin system for community extensions
- [ ] Webhook support for external integrations
- [ ] Savings goals with visual progress tracking

---

## 15. Getting Started (Day 1 Checklist)

```bash
# 1. Initialize the monorepo
mkdir moneyvault && cd moneyvault
git init

# 2. Scaffold the Go backend
mkdir -p backend/cmd/server backend/internal/{config,handlers,services,repositories,models,middleware,crypto}
cd backend && go mod init github.com/yourusername/moneyvault && cd ..

# 3. Scaffold the React frontend
npm create vite@latest frontend -- --template react-ts
cd frontend && npm install tailwindcss @tailwindcss/vite && cd ..

# 4. Create Docker Compose
touch docker-compose.yml docker-compose.dev.yml .env.example

# 5. Set up the database
mkdir -p backend/migrations
# Create initial migration: users, accounts, transactions, categories

# 6. Start building Phase 1!
docker compose -f docker-compose.dev.yml up
```

---

*MoneyVault — Your money. Your server. Your rules.*
