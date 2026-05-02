# Crypto Arbitrage Dashboard — Design Spec

**Date:** 2026-05-02  
**Status:** Approved

---

## Context

Arbitrage traders need to monitor price differences for the same asset across multiple exchanges simultaneously. Manually checking each exchange is slow; profitable windows are typically measured in seconds. This dashboard aggregates real-time prices from Binance, Coinbase, and Kraken into a single view, highlights cross-exchange spread opportunities above a configurable threshold, and delivers alerts in-app and via Telegram.

---

## Requirements

- **Exchanges:** Binance, Coinbase, Kraken (WebSocket price feeds, no API keys needed for market data)
- **Pairs:** Major pairs by volume — BTC/USDT, ETH/USDT, SOL/USDT, BNB/USDT, XRP/USDT, ADA/USDT, DOGE/USDT, MATIC/POL/USDT, LTC/USDT, DOT/USDT. Not every pair is listed on every exchange (e.g., BNB is Binance-native, absent on Coinbase/Kraken). Missing pairs show `—` in the price grid; spread calculation is skipped for that (symbol, exchange) combo.
- **Purpose:** Monitor only — no trade execution
- **Alerts:** In-app (visual toast + optional sound) + Telegram bot
- **Users:** Multi-user with email/password auth; each user has their own threshold and Telegram config
- **Stack:** Go backend, Vue.js + Tailwind CSS frontend (light theme)

---

## Architecture

A single Go binary connects to all exchange WebSocket feeds, aggregates prices, calculates spreads, and fans out real-time updates to authenticated browser clients over WebSocket. PostgreSQL stores users, settings, and alert history.

```
Exchange WebSocket feeds (Binance, Coinbase, Kraken)
         │
         ▼
┌─────────────────────────────────────────────────────────┐
│                    Go Backend (:8080)                   │
│                                                         │
│  Exchange Connectors → Price Aggregator → Alert Engine  │
│                              │                          │
│                       WebSocket Hub ──────────────────► Browser clients
│                                                         │
│  REST API (/api/auth, /api/settings, /api/history)      │
│                                                         │
│  PostgreSQL (users, user_settings, alert_history)       │
└─────────────────────────────────────────────────────────┘
         │ WebSocket (/ws)
         ▼
┌────────────────────────┐
│  Vue.js + Tailwind     │
│  Dashboard (:5173)     │
└────────────────────────┘
```

---

## Backend

### Project layout

```
backend/
├── cmd/server/main.go
├── internal/
│   ├── exchange/
│   │   ├── binance.go          # WS connector + normalizer
│   │   ├── coinbase.go
│   │   ├── kraken.go
│   │   └── types.go            # PriceTick, common interfaces
│   ├── aggregator/
│   │   ├── aggregator.go       # collects ticks, computes spreads
│   │   └── spread.go           # SpreadOpportunity type + calculation
│   ├── alert/
│   │   ├── engine.go           # evaluates spreads vs per-user thresholds
│   │   └── telegram.go         # Telegram Bot API delivery
│   ├── hub/
│   │   └── hub.go              # WebSocket hub, authenticated client registry
│   ├── api/
│   │   ├── auth.go             # POST /api/auth/register, /api/auth/login
│   │   ├── settings.go         # GET/PUT /api/settings
│   │   └── history.go          # GET /api/history
│   └── db/
│       ├── db.go               # connection + migrations
│       └── queries.go          # CRUD for users, settings, alert_history
├── migrations/
│   ├── 001_users.sql
│   ├── 002_user_settings.sql
│   └── 003_alert_history.sql
└── go.mod
```

### Key types

```go
type PriceTick struct {
    Exchange  string
    Symbol    string    // "BTC/USDT"
    Bid       float64
    Ask       float64
    Timestamp time.Time
}

type SpreadOpportunity struct {
    Symbol     string
    BuyAt      string  // exchange name
    SellAt     string
    BuyPrice   float64
    SellPrice  float64
    SpreadPct  float64
    DetectedAt time.Time
}
```

### Database schema

**users**
| Column | Type |
|---|---|
| id | UUID PK |
| email | TEXT UNIQUE |
| password_hash | TEXT |
| created_at | TIMESTAMPTZ |

**user_settings**
| Column | Type | Default |
|---|---|---|
| user_id | UUID FK |
| threshold_pct | FLOAT | 0.1 |
| telegram_bot_token | TEXT | — |
| telegram_chat_id | TEXT | — |
| in_app_alerts | BOOL | true |
| alert_sound | BOOL | true |

**alert_history**
| Column | Type |
|---|---|
| id | UUID PK |
| user_id | UUID FK |
| symbol | TEXT |
| buy_exchange | TEXT |
| sell_exchange | TEXT |
| spread_pct | FLOAT |
| buy_price | FLOAT |
| sell_price | FLOAT |
| fired_at | TIMESTAMPTZ |

### Exchange WebSocket endpoints

| Exchange | URL |
|---|---|
| Binance | `wss://stream.binance.com:9443/ws/<symbol>@bookTicker` |
| Coinbase | `wss://advanced-trade-ws.coinbase.com/ws` |
| Kraken | `wss://ws.kraken.com` |

No API keys required for public market data on any of the three.

### Alert engine logic

1. After each price tick, compute spread for every (symbol, buy_exchange, sell_exchange) triple.
2. For each connected user: if `spread_pct >= user.threshold_pct` and the opportunity was not already active → fire alert.
3. Debounce: re-alert the same opportunity only after it dips below threshold and recovers (prevents spam on volatile pairs).
4. In-app: push `{type:"alert", ...}` message over the user's WebSocket connection.
5. Telegram: POST to `api.telegram.org/bot{token}/sendMessage` with opportunity details.
6. Log to `alert_history`.

### Auth

- Email + password, bcrypt hashing.
- JWT (7-day expiry) returned on login/register.
- Token passed as `Authorization: Bearer <token>` on REST; as `?token=<jwt>` on WebSocket upgrade.

---

## Frontend

### Project layout

```
frontend/
├── src/
│   ├── main.js
│   ├── App.vue
│   ├── router/index.js
│   ├── stores/
│   │   ├── auth.js             # Pinia: JWT, user profile, login/logout
│   │   └── prices.js           # Pinia: ticks, opportunities, alert queue
│   ├── composables/
│   │   └── useWebSocket.js     # WS connection lifecycle + reconnect
│   ├── views/
│   │   ├── LoginView.vue
│   │   ├── DashboardView.vue   # stat cards + opportunities table
│   │   ├── AllPricesView.vue   # full pair × exchange price grid
│   │   ├── AlertHistoryView.vue
│   │   └── SettingsView.vue    # threshold, Telegram config
│   └── components/
│       ├── StatCard.vue         # summary metric card
│       ├── OpportunityTable.vue # live arbitrage opportunities
│       ├── PriceGrid.vue        # full price table (all pairs × exchanges)
│       ├── AlertToast.vue       # in-app notification
│       └── NavBar.vue
├── tailwind.config.js
└── vite.config.js
```

### Dashboard layout (light theme)

**Nav:** Logo | Dashboard | All Prices | Alert History | Settings | Logout

**Dashboard view:**
- Row of 3 stat cards: Active Opportunities / Avg Spread / Alerts Today
- `OpportunityTable`: columns — Pair | Buy At | Sell At | Spread % | Detected
- Rows highlighted green (`bg-green-50 border-l-4 border-green-400`) for active opportunities
- Empty state: "No opportunities above your threshold right now."

**All Prices view:**
- Table: Pair | Binance Bid/Ask | Coinbase Bid/Ask | Kraken Bid/Ask | Best Spread %
- Sortable by spread %, updates in real-time

**Settings view:**
- Threshold % input (number, step 0.01)
- Telegram Bot Token + Chat ID fields
- In-app alerts toggle, sound toggle
- Save button → PUT /api/settings

### State management (Pinia)

`prices` store:
- `ticks`: `Map<"exchange:symbol", PriceTick>` — latest tick per feed
- `opportunities`: computed — all (symbol, buyEx, sellEx) combos where spread ≥ `auth.settings.threshold_pct` (reads from `auth` store)
- `alertQueue`: incoming alert messages from WS, consumed by `AlertToast`

`useWebSocket` composable:
- Opens `ws://localhost:8080/ws?token=<jwt>`
- Dispatches `tick` events to `prices` store, `alert` events to `alertQueue`
- Reconnects with exponential backoff on disconnect

---

## Verification

```bash
# Start backend (requires PostgreSQL running)
cd backend && go run ./cmd/server

# Start frontend
cd frontend && npm run dev
```

End-to-end test steps:
1. Register + login → JWT stored, redirect to Dashboard
2. WebSocket connects → stat cards show `0 active opportunities`
3. All Prices tab populates with live prices within a few seconds
4. Set threshold to 0.01% in Settings → an opportunity appears in Dashboard table
5. In-app toast fires with pair + spread details
6. Configure Telegram bot token + chat ID → verify Telegram message received
7. Check Alert History → past alerts logged with timestamp, pair, spread %, exchanges
8. Raise threshold to 0.5% → opportunities table clears to empty state
