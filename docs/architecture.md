# Crypto arbitrage dashboard — architecture

This document reflects the repository layout and packages. The design spec also references `backend/cmd/server/main.go`; that entrypoint may not be present in every checkout—internal packages describe the intended runtime.

---

## High-level system

```mermaid
flowchart TB
  subgraph external["External systems"]
    BN[Binance WebSocket]
    CB[Coinbase WebSocket]
    KR[Kraken WebSocket]
    CDX[CoinDCX REST]
    CSW[CoinSwitch REST]
    IR[Independent Reserve REST]
    PG[(PostgreSQL)]
    TG[Telegram Bot API]
  end

  subgraph backend["Go backend intended :8080"]
    EX[exchange connectors]
    AGG[aggregator]
    ALT[alert engine]
    HUB[WebSocket hub]
    API[REST + WS handlers]
    DBL[db: pool, migrations, queries]
  end

  subgraph frontend["Vue 3 + Vite dev :5173"]
    UI[Views + components]
    ST[Pinia: auth, prices]
    WS[useWebSocket composable]
  end

  BN --> EX
  CB --> EX
  KR --> EX
  CDX --> EX
  CSW --> EX
  IR --> EX
  EX --> AGG
  AGG --> ALT
  ALT --> HUB
  ALT --> TG
  ALT --> DBL
  API --> DBL
  HUB --> WS
  API --> UI
  DBL --> PG
  WS -->|proxy /ws| HUB
  UI -->|proxy /api| API
```

---

## Exchange transports

Binance, Coinbase, and Kraken stream live book-ticker frames over public WebSocket. CoinDCX, CoinSwitch, and Independent Reserve do not expose a usable public WebSocket for this dataset, so their connectors poll REST endpoints (CoinDCX: bulk `/exchange/ticker` every 3s; CoinSwitch: bulk `/trade/api/v2/24hr/all-pairs/ticker?exchange=coinswitchx` every 3s; Independent Reserve: per-pair `GetMarketSummary` every 5s, fan-out one goroutine per pair) and emit normalized `PriceTick`s onto the same channel. The aggregator and downstream packages are transport-agnostic.

CoinDCX contributes USDT and INR pairs; CoinSwitch contributes INR pairs; Independent Reserve contributes SGD pairs. Quote currency is encoded in the symbol (e.g. `BTC/USDT`, `BTC/INR`, `BTC/SGD`), so spread comparison naturally stays within a single quote currency.

---

## Backend packages (`backend/internal/`)

```mermaid
flowchart LR
  subgraph exchange["exchange"]
    T[PriceTick + TickHandler]
    B[Binance]
    C[Coinbase]
    K[Kraken]
    D[CoinDCX]
    CS[CoinSwitch]
    IRX[IndepReserve]
  end

  subgraph agg["aggregator"]
    A[Aggregator: latest ticks]
    S[spread: SpreadOpportunity]
  end

  subgraph alert_pkg["alert"]
    E[Engine: threshold per user]
    TE[telegram.go]
  end

  subgraph hub_pkg["hub"]
    H[Register / Broadcast / SendToUser]
  end

  subgraph api_pkg["api"]
    AUTH[JWT + register/login]
    SET[settings GET/PUT]
    HIS[alert history GET]
    WSH[WebSocket: snapshot + pings]
    MW[JWTMiddleware]
  end

  subgraph db_pkg["db"]
    Q[Queries]
    M[Migrate + Connect]
  end

  B --> T
  C --> T
  K --> T
  D --> T
  CS --> T
  IRX --> T
  T --> A
  A --> S
  S --> E
  E --> H
  E --> TE
  E --> Q
  AUTH --> Q
  SET --> Q
  HIS --> Q
  WSH --> H
  WSH --> A
  MW --> SET
  MW --> HIS
  M --> Q
```

---

## Frontend (`frontend/src/`)

```mermaid
flowchart TB
  main[main.js: App + Pinia + Router]
  App[App.vue: NavBar, router-view, AlertToast]
  R[Router: login, dashboard, prices, history, settings]
  AS[stores/auth.js]
  PS[stores/prices.js]
  UWS[useWebSocket.js]

  main --> App
  main --> R
  App --> AS
  App --> UWS
  UWS -->|JWT query param| PS
  AS -->|Bearer /api| R
```

---

## Runtime flows

### Live prices and WebSocket snapshot

```mermaid
sequenceDiagram
  participant X as Exchanges WS
  participant E as exchange package
  participant A as Aggregator
  participant W as WSHandler
  participant H as Hub
  participant B as Browser

  X->>E: raw book ticker frames
  E->>A: PriceTick normalized
  Note over W,B: On connect
  B->>W: GET /ws?token=JWT
  W->>H: Register client
  W->>A: GetTicks + ComputeOpportunities
  W->>B: snapshot ticks + opportunities
```

### Alerts

```mermaid
sequenceDiagram
  participant A as Aggregator
  participant E as Alert engine
  participant Q as db.Queries
  participant H as Hub
  participant T as Telegram
  participant B as Browser

  Note over E: Evaluate when opportunities list updates
  E->>Q: GetAllSettings
  E->>E: compare SpreadPct vs threshold
  alt in-app enabled
    E->>H: SendToUser alert JSON
    H->>B: type alert
  end
  alt Telegram configured
    E->>T: SendTelegram
  end
  E->>Q: InsertAlert history
```

---

## REST and auth

- **Auth:** `POST` register/login → bcrypt + JWT (`internal/api/auth.go`).
- **Protected routes:** JWT from `Authorization: Bearer` or `?token=` (`JWTMiddleware`).
- **Settings / history:** `internal/api/settings.go`, `internal/api/history.go` → `db.Queries`.

## Local development proxy

`frontend/vite.config.js` proxies `/api` and `/ws` to `http://localhost:8080` and `ws://localhost:8080`.

## Related docs

- [Design spec](superpowers/specs/2026-05-02-crypto-arbitrage-dashboard-design.md)
