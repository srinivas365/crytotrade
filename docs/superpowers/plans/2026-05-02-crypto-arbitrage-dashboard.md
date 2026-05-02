# Crypto Arbitrage Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a real-time crypto arbitrage dashboard that monitors major pair prices across Binance, Coinbase, and Kraken, highlights spread opportunities, and delivers in-app + Telegram alerts to authenticated users.

**Architecture:** Go monolith connects to exchange WebSocket feeds, aggregates prices, evaluates per-user thresholds in the alert engine, and fans out 500ms snapshots to browsers via WebSocket hub. Vue 3 + Tailwind frontend (light theme) shows stat cards, opportunity table, full price grid, alert history, and settings.

**Tech Stack:** Go 1.23, chi v5, pgx/v5, gorilla/websocket, golang-jwt/jwt/v5, bcrypt; Vue 3, Vite, Pinia, Vue Router 4, Tailwind CSS 3, Vitest

---

## File Map

```
cryptotrade/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/config.go
│   │   ├── db/
│   │   │   ├── db.go
│   │   │   └── queries.go
│   │   ├── exchange/
│   │   │   ├── types.go
│   │   │   ├── binance.go
│   │   │   ├── coinbase.go
│   │   │   └── kraken.go
│   │   ├── aggregator/
│   │   │   ├── spread.go
│   │   │   └── aggregator.go
│   │   ├── alert/
│   │   │   ├── engine.go
│   │   │   └── telegram.go
│   │   ├── hub/hub.go
│   │   └── api/
│   │       ├── auth.go
│   │       ├── settings.go
│   │       ├── history.go
│   │       └── ws.go
│   ├── migrations/
│   │   ├── 001_users.sql
│   │   ├── 002_user_settings.sql
│   │   └── 003_alert_history.sql
│   └── go.mod
└── frontend/
    ├── src/
    │   ├── main.js
    │   ├── App.vue
    │   ├── router/index.js
    │   ├── stores/auth.js
    │   ├── stores/prices.js
    │   ├── composables/useWebSocket.js
    │   ├── views/
    │   │   ├── LoginView.vue
    │   │   ├── DashboardView.vue
    │   │   ├── AllPricesView.vue
    │   │   ├── AlertHistoryView.vue
    │   │   └── SettingsView.vue
    │   └── components/
    │       ├── NavBar.vue
    │       ├── StatCard.vue
    │       ├── OpportunityTable.vue
    │       ├── PriceGrid.vue
    │       └── AlertToast.vue
    ├── tailwind.config.js
    └── vite.config.js
```

---

## Task 1: Go module + project scaffold

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go` (stub)

- [ ] **Step 1: Create directory structure**

```bash
cd /Users/mannemsrinivas/projects/cryptotrade
mkdir -p backend/cmd/server
mkdir -p backend/internal/{config,db,exchange,aggregator,alert,hub,api}
mkdir -p backend/migrations
```

- [ ] **Step 2: Initialise Go module**

```bash
cd backend
go mod init github.com/cryptotrade/app
```

- [ ] **Step 3: Add dependencies**

```bash
go get github.com/go-chi/chi/v5@v5.1.0
go get github.com/gorilla/websocket@v1.5.3
go get github.com/jackc/pgx/v5@v5.7.1
go get github.com/golang-jwt/jwt/v5@v5.2.1
go get golang.org/x/crypto@v0.28.0
```

- [ ] **Step 4: Write stub main.go**

`backend/cmd/server/main.go`:
```go
package main

import "fmt"

func main() {
	fmt.Println("cryptotrade starting")
}
```

- [ ] **Step 5: Verify it compiles**

```bash
go build ./...
```
Expected: no output (success).

- [ ] **Step 6: Commit**

```bash
cd /Users/mannemsrinivas/projects/cryptotrade
git init
git add backend/
git commit -m "feat: initialise Go module and project scaffold"
```

---

## Task 2: Config

**Files:**
- Create: `backend/internal/config/config.go`

- [ ] **Step 1: Write config.go**

```go
package config

import "os"

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
}

func Load() Config {
	return Config{
		DatabaseURL: env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/cryptotrade?sslmode=disable"),
		JWTSecret:   env("JWT_SECRET", "change-me-in-production"),
		Port:        env("PORT", "8080"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Step 2: Verify**

```bash
cd backend && go build ./...
```
Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/config/
git commit -m "feat: add config loader from environment"
```

---

## Task 3: Database migrations

**Files:**
- Create: `backend/migrations/001_users.sql`
- Create: `backend/migrations/002_user_settings.sql`
- Create: `backend/migrations/003_alert_history.sql`

- [ ] **Step 1: Write 001_users.sql**

```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

- [ ] **Step 2: Write 002_user_settings.sql**

```sql
CREATE TABLE user_settings (
    user_id           UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    threshold_pct     FLOAT NOT NULL DEFAULT 0.1,
    telegram_bot_token TEXT NOT NULL DEFAULT '',
    telegram_chat_id  TEXT NOT NULL DEFAULT '',
    in_app_alerts     BOOL NOT NULL DEFAULT true,
    alert_sound       BOOL NOT NULL DEFAULT true
);
```

- [ ] **Step 3: Write 003_alert_history.sql**

```sql
CREATE TABLE alert_history (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    symbol        TEXT NOT NULL,
    buy_exchange  TEXT NOT NULL,
    sell_exchange TEXT NOT NULL,
    spread_pct    FLOAT NOT NULL,
    buy_price     FLOAT NOT NULL,
    sell_price    FLOAT NOT NULL,
    fired_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX alert_history_user_fired_idx ON alert_history(user_id, fired_at DESC);
```

- [ ] **Step 4: Commit**

```bash
git add backend/migrations/
git commit -m "feat: add database migration SQL files"
```

---

## Task 4: Database connection + migration runner

**Files:**
- Create: `backend/internal/db/db.go`

- [ ] **Step 1: Write db.go**

```go
package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename   TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		var applied bool
		if err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename=$1)`, name,
		).Scan(&applied); err != nil {
			return fmt.Errorf("check %s: %w", name, err)
		}
		if applied {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(content)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("apply %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations(filename) VALUES($1)`, name); err != nil {
			tx.Rollback(ctx)
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 2: Verify**

```bash
cd backend && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/db/db.go
git commit -m "feat: add database connection and migration runner"
```

---

## Task 5: Database queries

**Files:**
- Create: `backend/internal/db/queries.go`

- [ ] **Step 1: Write queries.go**

```go
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type UserSettings struct {
	UserID           string  `json:"user_id"`
	ThresholdPct     float64 `json:"threshold_pct"`
	TelegramBotToken string  `json:"telegram_bot_token"`
	TelegramChatID   string  `json:"telegram_chat_id"`
	InAppAlerts      bool    `json:"in_app_alerts"`
	AlertSound       bool    `json:"alert_sound"`
}

type AlertRecord struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Symbol       string    `json:"symbol"`
	BuyExchange  string    `json:"buy_exchange"`
	SellExchange string    `json:"sell_exchange"`
	SpreadPct    float64   `json:"spread_pct"`
	BuyPrice     float64   `json:"buy_price"`
	SellPrice    float64   `json:"sell_price"`
	FiredAt      time.Time `json:"fired_at"`
}

type Queries struct{ pool *pgxpool.Pool }

func NewQueries(pool *pgxpool.Pool) *Queries { return &Queries{pool: pool} }

func (q *Queries) CreateUser(ctx context.Context, email, hash string) (*User, error) {
	u := &User{}
	if err := q.pool.QueryRow(ctx,
		`INSERT INTO users(email,password_hash) VALUES($1,$2)
		 RETURNING id,email,password_hash,created_at`,
		email, hash,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	if _, err := q.pool.Exec(ctx,
		`INSERT INTO user_settings(user_id) VALUES($1) ON CONFLICT DO NOTHING`, u.ID,
	); err != nil {
		return nil, fmt.Errorf("create settings: %w", err)
	}
	return u, nil
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	u := &User{}
	if err := q.pool.QueryRow(ctx,
		`SELECT id,email,password_hash,created_at FROM users WHERE email=$1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

func (q *Queries) GetSettings(ctx context.Context, userID string) (*UserSettings, error) {
	s := &UserSettings{}
	if err := q.pool.QueryRow(ctx,
		`SELECT user_id,threshold_pct,telegram_bot_token,telegram_chat_id,in_app_alerts,alert_sound
		 FROM user_settings WHERE user_id=$1`, userID,
	).Scan(&s.UserID, &s.ThresholdPct, &s.TelegramBotToken, &s.TelegramChatID, &s.InAppAlerts, &s.AlertSound); err != nil {
		return nil, err
	}
	return s, nil
}

func (q *Queries) UpsertSettings(ctx context.Context, s *UserSettings) error {
	_, err := q.pool.Exec(ctx,
		`INSERT INTO user_settings(user_id,threshold_pct,telegram_bot_token,telegram_chat_id,in_app_alerts,alert_sound)
		 VALUES($1,$2,$3,$4,$5,$6)
		 ON CONFLICT(user_id) DO UPDATE SET
		   threshold_pct=EXCLUDED.threshold_pct,
		   telegram_bot_token=EXCLUDED.telegram_bot_token,
		   telegram_chat_id=EXCLUDED.telegram_chat_id,
		   in_app_alerts=EXCLUDED.in_app_alerts,
		   alert_sound=EXCLUDED.alert_sound`,
		s.UserID, s.ThresholdPct, s.TelegramBotToken, s.TelegramChatID, s.InAppAlerts, s.AlertSound,
	)
	return err
}

func (q *Queries) GetAllSettings(ctx context.Context) ([]*UserSettings, error) {
	rows, err := q.pool.Query(ctx,
		`SELECT user_id,threshold_pct,telegram_bot_token,telegram_chat_id,in_app_alerts,alert_sound FROM user_settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*UserSettings
	for rows.Next() {
		s := &UserSettings{}
		if err := rows.Scan(&s.UserID, &s.ThresholdPct, &s.TelegramBotToken, &s.TelegramChatID, &s.InAppAlerts, &s.AlertSound); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (q *Queries) InsertAlert(ctx context.Context, r *AlertRecord) error {
	_, err := q.pool.Exec(ctx,
		`INSERT INTO alert_history(user_id,symbol,buy_exchange,sell_exchange,spread_pct,buy_price,sell_price)
		 VALUES($1,$2,$3,$4,$5,$6,$7)`,
		r.UserID, r.Symbol, r.BuyExchange, r.SellExchange, r.SpreadPct, r.BuyPrice, r.SellPrice,
	)
	return err
}

func (q *Queries) GetAlertHistory(ctx context.Context, userID string, limit int) ([]*AlertRecord, error) {
	rows, err := q.pool.Query(ctx,
		`SELECT id,user_id,symbol,buy_exchange,sell_exchange,spread_pct,buy_price,sell_price,fired_at
		 FROM alert_history WHERE user_id=$1 ORDER BY fired_at DESC LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*AlertRecord
	for rows.Next() {
		r := &AlertRecord{}
		if err := rows.Scan(&r.ID, &r.UserID, &r.Symbol, &r.BuyExchange, &r.SellExchange,
			&r.SpreadPct, &r.BuyPrice, &r.SellPrice, &r.FiredAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
```

- [ ] **Step 2: Verify**

```bash
cd backend && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/db/queries.go
git commit -m "feat: add database query functions for users, settings, and alerts"
```

---

## Task 6: Exchange types

**Files:**
- Create: `backend/internal/exchange/types.go`

- [ ] **Step 1: Write types.go**

```go
package exchange

import "time"

type PriceTick struct {
	Exchange  string    `json:"exchange"`
	Symbol    string    `json:"symbol"` // normalised: "BTC/USDT"
	Bid       float64   `json:"bid"`
	Ask       float64   `json:"ask"`
	Timestamp time.Time `json:"timestamp"`
}

// TickHandler receives normalised ticks from a connector.
type TickHandler func(PriceTick)
```

- [ ] **Step 2: Write exchange/types_test.go**

```go
package exchange

import (
	"testing"
	"time"
)

func TestPriceTick_fields(t *testing.T) {
	tick := PriceTick{
		Exchange:  "binance",
		Symbol:    "BTC/USDT",
		Bid:       63241.0,
		Ask:       63242.0,
		Timestamp: time.Now(),
	}
	if tick.Exchange != "binance" {
		t.Fatalf("expected binance, got %s", tick.Exchange)
	}
	if tick.Bid >= tick.Ask {
		t.Fatalf("bid should be less than ask")
	}
}
```

- [ ] **Step 3: Run test**

```bash
cd backend && go test ./internal/exchange/... -v
```
Expected: `PASS`

- [ ] **Step 4: Commit**

```bash
git add backend/internal/exchange/
git commit -m "feat: add exchange PriceTick type"
```

---

## Task 7: Binance connector

**Files:**
- Create: `backend/internal/exchange/binance.go`
- Create: `backend/internal/exchange/binance_test.go`

Symbol map — Binance uses `BTCUSDT`, we normalise to `BTC/USDT`.

- [ ] **Step 1: Write binance_test.go**

```go
package exchange

import (
	"encoding/json"
	"testing"
)

func TestBinance_parseTick(t *testing.T) {
	raw := `{"stream":"btcusdt@bookTicker","data":{"u":123,"s":"BTCUSDT","b":"63241.00","B":"0.5","a":"63242.50","A":"0.3"}}`
	b := &Binance{}
	tick, err := b.parseMessage([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if tick == nil {
		t.Fatal("expected tick, got nil")
	}
	if tick.Exchange != "binance" {
		t.Fatalf("exchange: got %s", tick.Exchange)
	}
	if tick.Symbol != "BTC/USDT" {
		t.Fatalf("symbol: got %s", tick.Symbol)
	}
	if tick.Bid != 63241.00 {
		t.Fatalf("bid: got %f", tick.Bid)
	}
	if tick.Ask != 63242.50 {
		t.Fatalf("ask: got %f", tick.Ask)
	}
	_ = json.Marshal(tick)
}

func TestBinance_parseTickUnknownSymbol(t *testing.T) {
	raw := `{"stream":"xyzabc@bookTicker","data":{"s":"XYZABC","b":"1.0","a":"1.1"}}`
	b := &Binance{}
	tick, err := b.parseMessage([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if tick != nil {
		t.Fatal("expected nil for unknown symbol")
	}
}
```

- [ ] **Step 2: Run test to confirm failure**

```bash
cd backend && go test ./internal/exchange/... -run TestBinance -v
```
Expected: compile error (Binance type not defined yet).

- [ ] **Step 3: Write binance.go**

```go
package exchange

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var binanceSymbols = map[string]string{
	"BTCUSDT":  "BTC/USDT",
	"ETHUSDT":  "ETH/USDT",
	"SOLUSDT":  "SOL/USDT",
	"BNBUSDT":  "BNB/USDT",
	"XRPUSDT":  "XRP/USDT",
	"ADAUSDT":  "ADA/USDT",
	"DOGEUSDT": "DOGE/USDT",
	"MATICUSDT":"MATIC/USDT",
	"LTCUSDT":  "LTC/USDT",
	"DOTUSDT":  "DOT/USDT",
}

type Binance struct{}

type binanceStreamMsg struct {
	Stream string `json:"stream"`
	Data   struct {
		Symbol string `json:"s"`
		Bid    string `json:"b"`
		Ask    string `json:"a"`
	} `json:"data"`
}

func (b *Binance) parseMessage(raw []byte) (*PriceTick, error) {
	var msg binanceStreamMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, err
	}
	symbol, ok := binanceSymbols[msg.Data.Symbol]
	if !ok {
		return nil, nil
	}
	bid, _ := strconv.ParseFloat(msg.Data.Bid, 64)
	ask, _ := strconv.ParseFloat(msg.Data.Ask, 64)
	return &PriceTick{
		Exchange:  "binance",
		Symbol:    symbol,
		Bid:       bid,
		Ask:       ask,
		Timestamp: time.Now(),
	}, nil
}

func (b *Binance) Connect(ctx context.Context, out chan<- PriceTick) {
	keys := make([]string, 0, len(binanceSymbols))
	for k := range binanceSymbols {
		keys = append(keys, strings.ToLower(k)+"@bookTicker")
	}
	url := "wss://stream.binance.com:9443/stream?streams=" + strings.Join(keys, "/")

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
		if err != nil {
			log.Printf("binance dial: %v — retrying in 5s", err)
			time.Sleep(5 * time.Second)
			continue
		}
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("binance read: %v", err)
				conn.Close()
				break
			}
			tick, err := b.parseMessage(msg)
			if err != nil || tick == nil {
				continue
			}
			select {
			case out <- *tick:
			default:
			}
		}
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd backend && go test ./internal/exchange/... -run TestBinance -v
```
Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add backend/internal/exchange/binance.go backend/internal/exchange/binance_test.go
git commit -m "feat: add Binance bookTicker WebSocket connector"
```

---

## Task 8: Coinbase connector

**Files:**
- Create: `backend/internal/exchange/coinbase.go`
- Create: `backend/internal/exchange/coinbase_test.go`

Coinbase uses USD (not USDT). We treat USD ≈ USDT for spread comparison.

- [ ] **Step 1: Write coinbase_test.go**

```go
package exchange

import "testing"

func TestCoinbase_parseTick(t *testing.T) {
	// Coinbase Advanced Trade ticker update format
	raw := `{"channel":"ticker","events":[{"type":"update","tickers":[{"product_id":"BTC-USD","best_bid":"63198.00","best_ask":"63199.50"}]}]}`
	c := &Coinbase{}
	ticks, err := c.parseMessage([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(ticks) != 1 {
		t.Fatalf("expected 1 tick, got %d", len(ticks))
	}
	tick := ticks[0]
	if tick.Exchange != "coinbase" {
		t.Fatalf("exchange: got %s", tick.Exchange)
	}
	if tick.Symbol != "BTC/USDT" {
		t.Fatalf("symbol: got %s", tick.Symbol)
	}
	if tick.Bid != 63198.00 {
		t.Fatalf("bid: got %f", tick.Bid)
	}
	if tick.Ask != 63199.50 {
		t.Fatalf("ask: got %f", tick.Ask)
	}
}

func TestCoinbase_parseTickUnknownProduct(t *testing.T) {
	raw := `{"channel":"ticker","events":[{"type":"update","tickers":[{"product_id":"BNB-USD","best_bid":"574.0","best_ask":"574.5"}]}]}`
	c := &Coinbase{}
	ticks, _ := c.parseMessage([]byte(raw))
	if len(ticks) != 0 {
		t.Fatal("BNB-USD not supported on Coinbase, expected no ticks")
	}
}
```

- [ ] **Step 2: Run test to confirm failure**

```bash
cd backend && go test ./internal/exchange/... -run TestCoinbase -v
```
Expected: compile error.

- [ ] **Step 3: Write coinbase.go**

```go
package exchange

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var coinbaseSymbols = map[string]string{
	"BTC-USD":  "BTC/USDT",
	"ETH-USD":  "ETH/USDT",
	"SOL-USD":  "SOL/USDT",
	"XRP-USD":  "XRP/USDT",
	"ADA-USD":  "ADA/USDT",
	"DOGE-USD": "DOGE/USDT",
	"LTC-USD":  "LTC/USDT",
	"DOT-USD":  "DOT/USDT",
}

type Coinbase struct{}

type coinbaseMsg struct {
	Channel string `json:"channel"`
	Events  []struct {
		Type    string `json:"type"`
		Tickers []struct {
			ProductID string `json:"product_id"`
			BestBid   string `json:"best_bid"`
			BestAsk   string `json:"best_ask"`
		} `json:"tickers"`
	} `json:"events"`
}

func (c *Coinbase) parseMessage(raw []byte) ([]PriceTick, error) {
	var msg coinbaseMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, err
	}
	if msg.Channel != "ticker" {
		return nil, nil
	}
	var ticks []PriceTick
	for _, ev := range msg.Events {
		for _, t := range ev.Tickers {
			symbol, ok := coinbaseSymbols[t.ProductID]
			if !ok {
				continue
			}
			bid, _ := strconv.ParseFloat(t.BestBid, 64)
			ask, _ := strconv.ParseFloat(t.BestAsk, 64)
			if bid == 0 && ask == 0 {
				continue
			}
			ticks = append(ticks, PriceTick{
				Exchange:  "coinbase",
				Symbol:    symbol,
				Bid:       bid,
				Ask:       ask,
				Timestamp: time.Now(),
			})
		}
	}
	return ticks, nil
}

func (c *Coinbase) Connect(ctx context.Context, out chan<- PriceTick) {
	url := "wss://advanced-trade-ws.coinbase.com/ws"
	products := make([]string, 0, len(coinbaseSymbols))
	for k := range coinbaseSymbols {
		products = append(products, k)
	}
	subscribe, _ := json.Marshal(map[string]interface{}{
		"type":        "subscribe",
		"product_ids": products,
		"channel":     "ticker",
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
		if err != nil {
			log.Printf("coinbase dial: %v — retrying in 5s", err)
			time.Sleep(5 * time.Second)
			continue
		}
		conn.WriteMessage(websocket.TextMessage, subscribe)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("coinbase read: %v", err)
				conn.Close()
				break
			}
			ticks, err := c.parseMessage(msg)
			if err != nil || len(ticks) == 0 {
				continue
			}
			for _, tick := range ticks {
				select {
				case out <- tick:
				default:
				}
			}
		}
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd backend && go test ./internal/exchange/... -run TestCoinbase -v
```
Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add backend/internal/exchange/coinbase.go backend/internal/exchange/coinbase_test.go
git commit -m "feat: add Coinbase Advanced Trade WebSocket connector"
```

---

## Task 9: Kraken connector

**Files:**
- Create: `backend/internal/exchange/kraken.go`
- Create: `backend/internal/exchange/kraken_test.go`

Kraken v2 WS at `wss://ws.kraken.com/v2`. Uses `XBT/USDT` for Bitcoin → normalise to `BTC/USDT`.

- [ ] **Step 1: Write kraken_test.go**

```go
package exchange

import "testing"

func TestKraken_parseTick(t *testing.T) {
	raw := `{"channel":"ticker","type":"update","data":[{"symbol":"XBT/USDT","bid":63310.0,"ask":63311.5}]}`
	k := &Kraken{}
	ticks, err := k.parseMessage([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(ticks) != 1 {
		t.Fatalf("expected 1 tick, got %d", len(ticks))
	}
	tick := ticks[0]
	if tick.Exchange != "kraken" {
		t.Fatalf("exchange: got %s", tick.Exchange)
	}
	if tick.Symbol != "BTC/USDT" {
		t.Fatalf("symbol: got %s, want BTC/USDT", tick.Symbol)
	}
	if tick.Bid != 63310.0 {
		t.Fatalf("bid: got %f", tick.Bid)
	}
}

func TestKraken_parseTickUnknownSymbol(t *testing.T) {
	raw := `{"channel":"ticker","type":"update","data":[{"symbol":"SHIB/USDT","bid":0.00001,"ask":0.00002}]}`
	k := &Kraken{}
	ticks, _ := k.parseMessage([]byte(raw))
	if len(ticks) != 0 {
		t.Fatal("expected no ticks for unsupported symbol")
	}
}
```

- [ ] **Step 2: Run test to confirm failure**

```bash
cd backend && go test ./internal/exchange/... -run TestKraken -v
```
Expected: compile error.

- [ ] **Step 3: Write kraken.go**

```go
package exchange

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var krakenSymbols = map[string]string{
	"XBT/USDT":  "BTC/USDT",
	"BTC/USDT":  "BTC/USDT",
	"ETH/USDT":  "ETH/USDT",
	"SOL/USDT":  "SOL/USDT",
	"XRP/USDT":  "XRP/USDT",
	"ADA/USDT":  "ADA/USDT",
	"DOGE/USDT": "DOGE/USDT",
	"LTC/USDT":  "LTC/USDT",
	"DOT/USDT":  "DOT/USDT",
}

type Kraken struct{}

type krakenMsg struct {
	Channel string `json:"channel"`
	Type    string `json:"type"`
	Data    []struct {
		Symbol string  `json:"symbol"`
		Bid    float64 `json:"bid"`
		Ask    float64 `json:"ask"`
	} `json:"data"`
}

func (k *Kraken) parseMessage(raw []byte) ([]PriceTick, error) {
	var msg krakenMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, err
	}
	if msg.Channel != "ticker" {
		return nil, nil
	}
	var ticks []PriceTick
	for _, d := range msg.Data {
		symbol, ok := krakenSymbols[d.Symbol]
		if !ok {
			continue
		}
		ticks = append(ticks, PriceTick{
			Exchange:  "kraken",
			Symbol:    symbol,
			Bid:       d.Bid,
			Ask:       d.Ask,
			Timestamp: time.Now(),
		})
	}
	return ticks, nil
}

func (k *Kraken) Connect(ctx context.Context, out chan<- PriceTick) {
	url := "wss://ws.kraken.com/v2"
	pairs := make([]string, 0, len(krakenSymbols))
	seen := map[string]bool{}
	for _, v := range krakenSymbols {
		if !seen[v] {
			seen[v] = true
			pairs = append(pairs, v)
		}
	}
	subscribe, _ := json.Marshal(map[string]interface{}{
		"method": "subscribe",
		"params": map[string]interface{}{
			"channel": "ticker",
			"symbol":  pairs,
		},
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
		if err != nil {
			log.Printf("kraken dial: %v — retrying in 5s", err)
			time.Sleep(5 * time.Second)
			continue
		}
		conn.WriteMessage(websocket.TextMessage, subscribe)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("kraken read: %v", err)
				conn.Close()
				break
			}
			ticks, err := k.parseMessage(msg)
			if err != nil || len(ticks) == 0 {
				continue
			}
			for _, tick := range ticks {
				select {
				case out <- tick:
				default:
				}
			}
		}
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd backend && go test ./internal/exchange/... -v
```
Expected: all `PASS`

- [ ] **Step 5: Commit**

```bash
git add backend/internal/exchange/kraken.go backend/internal/exchange/kraken_test.go
git commit -m "feat: add Kraken v2 WebSocket connector"
```

---

## Task 10: Spread calculator + aggregator

**Files:**
- Create: `backend/internal/aggregator/spread.go`
- Create: `backend/internal/aggregator/spread_test.go`
- Create: `backend/internal/aggregator/aggregator.go`

- [ ] **Step 1: Write spread_test.go**

```go
package aggregator

import (
	"testing"
	"time"

	"github.com/cryptotrade/app/internal/exchange"
)

func makeTick(exch, symbol string, bid, ask float64) exchange.PriceTick {
	return exchange.PriceTick{Exchange: exch, Symbol: symbol, Bid: bid, Ask: ask, Timestamp: time.Now()}
}

func TestCalculateSpread_positive(t *testing.T) {
	buy := makeTick("coinbase", "BTC/USDT", 63198, 63199)
	sell := makeTick("kraken", "BTC/USDT", 63310, 63311)
	opp := CalculateSpread("BTC/USDT", buy, sell)
	// buy at coinbase ask=63199, sell at kraken bid=63310
	expected := (63310 - 63199) / 63199 * 100
	if abs(opp.SpreadPct-expected) > 0.0001 {
		t.Fatalf("spread: got %f, want %f", opp.SpreadPct, expected)
	}
	if opp.BuyAt != "coinbase" || opp.SellAt != "kraken" {
		t.Fatalf("buy/sell: %s/%s", opp.BuyAt, opp.SellAt)
	}
}

func TestCalculateSpread_negative(t *testing.T) {
	buy := makeTick("kraken", "BTC/USDT", 63310, 63315)
	sell := makeTick("coinbase", "BTC/USDT", 63198, 63199)
	opp := CalculateSpread("BTC/USDT", buy, sell)
	if opp.SpreadPct >= 0 {
		t.Fatalf("expected negative spread, got %f", opp.SpreadPct)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
```

- [ ] **Step 2: Run test to confirm failure**

```bash
cd backend && go test ./internal/aggregator/... -run TestCalculateSpread -v
```
Expected: compile error.

- [ ] **Step 3: Write spread.go**

```go
package aggregator

import (
	"time"

	"github.com/cryptotrade/app/internal/exchange"
)

type SpreadOpportunity struct {
	Symbol     string    `json:"symbol"`
	BuyAt      string    `json:"buy_at"`
	SellAt     string    `json:"sell_at"`
	BuyPrice   float64   `json:"buy_price"`
	SellPrice  float64   `json:"sell_price"`
	SpreadPct  float64   `json:"spread_pct"`
	DetectedAt time.Time `json:"detected_at"`
}

// CalculateSpread returns the spread when buying at buyTick.Ask and selling at sellTick.Bid.
func CalculateSpread(symbol string, buyTick, sellTick exchange.PriceTick) SpreadOpportunity {
	buyPrice := buyTick.Ask
	sellPrice := sellTick.Bid
	spreadPct := 0.0
	if buyPrice > 0 {
		spreadPct = (sellPrice - buyPrice) / buyPrice * 100
	}
	return SpreadOpportunity{
		Symbol:     symbol,
		BuyAt:      buyTick.Exchange,
		SellAt:     sellTick.Exchange,
		BuyPrice:   buyPrice,
		SellPrice:  sellPrice,
		SpreadPct:  spreadPct,
		DetectedAt: time.Now(),
	}
}
```

- [ ] **Step 4: Run spread tests**

```bash
cd backend && go test ./internal/aggregator/... -run TestCalculateSpread -v
```
Expected: `PASS`

- [ ] **Step 5: Write aggregator.go**

```go
package aggregator

import (
	"sync"

	"github.com/cryptotrade/app/internal/exchange"
)

type Aggregator struct {
	mu    sync.RWMutex
	ticks map[string]exchange.PriceTick // "exchange:symbol" → latest tick
}

func New() *Aggregator {
	return &Aggregator{ticks: make(map[string]exchange.PriceTick)}
}

func (a *Aggregator) UpdateTick(tick exchange.PriceTick) {
	key := tick.Exchange + ":" + tick.Symbol
	a.mu.Lock()
	a.ticks[key] = tick
	a.mu.Unlock()
}

func (a *Aggregator) GetTicks() map[string]exchange.PriceTick {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make(map[string]exchange.PriceTick, len(a.ticks))
	for k, v := range a.ticks {
		out[k] = v
	}
	return out
}

func (a *Aggregator) ComputeOpportunities() []SpreadOpportunity {
	ticks := a.GetTicks()
	bySymbol := make(map[string][]exchange.PriceTick)
	for _, t := range ticks {
		bySymbol[t.Symbol] = append(bySymbol[t.Symbol], t)
	}
	var opps []SpreadOpportunity
	for symbol, ts := range bySymbol {
		for i := range ts {
			for j := range ts {
				if i == j {
					continue
				}
				opp := CalculateSpread(symbol, ts[i], ts[j])
				if opp.SpreadPct > 0 {
					opps = append(opps, opp)
				}
			}
		}
	}
	return opps
}
```

- [ ] **Step 6: Write aggregator test**

```go
// Append to spread_test.go:

func TestAggregator_ComputeOpportunities(t *testing.T) {
	agg := New()
	agg.UpdateTick(makeTick("coinbase", "BTC/USDT", 63198, 63199))
	agg.UpdateTick(makeTick("kraken",   "BTC/USDT", 63310, 63311))

	opps := agg.ComputeOpportunities()
	if len(opps) == 0 {
		t.Fatal("expected at least one opportunity")
	}
	best := opps[0]
	if best.SpreadPct <= 0 {
		t.Fatalf("spread must be positive, got %f", best.SpreadPct)
	}
	if best.BuyAt != "coinbase" || best.SellAt != "kraken" {
		t.Fatalf("buy/sell: %s/%s", best.BuyAt, best.SellAt)
	}
}
```

- [ ] **Step 7: Run all aggregator tests**

```bash
cd backend && go test ./internal/aggregator/... -v
```
Expected: all `PASS`

- [ ] **Step 8: Commit**

```bash
git add backend/internal/aggregator/
git commit -m "feat: add spread calculator and price aggregator"
```

---

## Task 11: WebSocket hub

**Files:**
- Create: `backend/internal/hub/hub.go`

- [ ] **Step 1: Write hub.go**

```go
package hub

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserID string
	Conn   *websocket.Conn
	Send   chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

func New() *Hub {
	return &Hub{clients: make(map[string]*Client)}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	h.clients[c.UserID] = c
	h.mu.Unlock()
}

func (h *Hub) Unregister(userID string) {
	h.mu.Lock()
	delete(h.clients, userID)
	h.mu.Unlock()
}

func (h *Hub) Broadcast(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients {
		select {
		case c.Send <- msg:
		default:
		}
	}
}

func (h *Hub) SendToUser(userID string, msg []byte) {
	h.mu.RLock()
	c, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return
	}
	select {
	case c.Send <- msg:
	default:
	}
}

func (h *Hub) ConnectedUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]string, 0, len(h.clients))
	for uid := range h.clients {
		out = append(out, uid)
	}
	return out
}
```

- [ ] **Step 2: Write hub_test.go**

Create `backend/internal/hub/hub_test.go`:
```go
package hub

import (
	"testing"
)

func TestHub_RegisterUnregister(t *testing.T) {
	h := New()
	c := &Client{UserID: "user1", Send: make(chan []byte, 1)}
	h.Register(c)
	if users := h.ConnectedUsers(); len(users) != 1 {
		t.Fatalf("expected 1 connected user, got %d", len(users))
	}
	h.Unregister("user1")
	if users := h.ConnectedUsers(); len(users) != 0 {
		t.Fatalf("expected 0 connected users, got %d", len(users))
	}
}

func TestHub_SendToUser(t *testing.T) {
	h := New()
	c := &Client{UserID: "user1", Send: make(chan []byte, 1)}
	h.Register(c)
	h.SendToUser("user1", []byte("hello"))
	msg := <-c.Send
	if string(msg) != "hello" {
		t.Fatalf("got %s", string(msg))
	}
}

func TestHub_Broadcast(t *testing.T) {
	h := New()
	c1 := &Client{UserID: "u1", Send: make(chan []byte, 1)}
	c2 := &Client{UserID: "u2", Send: make(chan []byte, 1)}
	h.Register(c1)
	h.Register(c2)
	h.Broadcast([]byte("ping"))
	if string(<-c1.Send) != "ping" || string(<-c2.Send) != "ping" {
		t.Fatal("broadcast failed")
	}
}
```

- [ ] **Step 3: Run hub tests**

```bash
cd backend && go test ./internal/hub/... -v
```
Expected: all `PASS`

- [ ] **Step 4: Commit**

```bash
git add backend/internal/hub/
git commit -m "feat: add WebSocket hub for client registry and broadcast"
```

---

## Task 12: Alert engine + Telegram

**Files:**
- Create: `backend/internal/alert/engine.go`
- Create: `backend/internal/alert/engine_test.go`
- Create: `backend/internal/alert/telegram.go`

- [ ] **Step 1: Write engine_test.go**

```go
package alert

import (
	"testing"
	"time"

	"github.com/cryptotrade/app/internal/aggregator"
)

func makeOpp(symbol, buyAt, sellAt string, spreadPct float64) aggregator.SpreadOpportunity {
	return aggregator.SpreadOpportunity{
		Symbol: symbol, BuyAt: buyAt, SellAt: sellAt,
		BuyPrice: 100, SellPrice: 100 + spreadPct,
		SpreadPct: spreadPct, DetectedAt: time.Now(),
	}
}

func TestEngine_firesOnNewOpportunity(t *testing.T) {
	fired := []string{}
	e := newTestEngine(func(userID string, opp aggregator.SpreadOpportunity) {
		fired = append(fired, userID+":"+opp.Symbol)
	})
	settings := testSettings("user1", 0.1)
	opp := makeOpp("BTC/USDT", "coinbase", "kraken", 0.18)

	e.evaluateForUser(settings, []aggregator.SpreadOpportunity{opp})
	if len(fired) != 1 {
		t.Fatalf("expected 1 alert fired, got %d", len(fired))
	}
}

func TestEngine_noDoubleFireWhileActive(t *testing.T) {
	fired := 0
	e := newTestEngine(func(_ string, _ aggregator.SpreadOpportunity) { fired++ })
	settings := testSettings("user1", 0.1)
	opp := makeOpp("BTC/USDT", "coinbase", "kraken", 0.18)

	e.evaluateForUser(settings, []aggregator.SpreadOpportunity{opp})
	e.evaluateForUser(settings, []aggregator.SpreadOpportunity{opp}) // same opp still active
	if fired != 1 {
		t.Fatalf("expected 1 alert, got %d (double-fired)", fired)
	}
}

func TestEngine_refireAfterRecovery(t *testing.T) {
	fired := 0
	e := newTestEngine(func(_ string, _ aggregator.SpreadOpportunity) { fired++ })
	settings := testSettings("user1", 0.1)
	opp := makeOpp("BTC/USDT", "coinbase", "kraken", 0.18)

	e.evaluateForUser(settings, []aggregator.SpreadOpportunity{opp}) // fires
	e.evaluateForUser(settings, []aggregator.SpreadOpportunity{})    // dips below threshold
	e.evaluateForUser(settings, []aggregator.SpreadOpportunity{opp}) // fires again on recovery
	if fired != 2 {
		t.Fatalf("expected 2 alerts (fire, recover, re-fire), got %d", fired)
	}
}
```

- [ ] **Step 2: Run test to confirm failure**

```bash
cd backend && go test ./internal/alert/... -v
```
Expected: compile error.

- [ ] **Step 3: Write engine.go**

```go
package alert

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cryptotrade/app/internal/aggregator"
	"github.com/cryptotrade/app/internal/db"
	"github.com/cryptotrade/app/internal/hub"
)

type Engine struct {
	queries    *db.Queries
	hub        *hub.Hub
	mu         sync.Mutex
	activeOpps map[string]map[string]bool // userID → oppKey → active
	onFire     func(string, aggregator.SpreadOpportunity) // for testing
}

func oppKey(opp aggregator.SpreadOpportunity) string {
	return fmt.Sprintf("%s|%s|%s", opp.Symbol, opp.BuyAt, opp.SellAt)
}

func New(queries *db.Queries, h *hub.Hub) *Engine {
	return &Engine{
		queries:    queries,
		hub:        h,
		activeOpps: make(map[string]map[string]bool),
	}
}

// newTestEngine creates an Engine with a fire callback (for unit tests without DB/hub).
func newTestEngine(onFire func(string, aggregator.SpreadOpportunity)) *Engine {
	return &Engine{
		activeOpps: make(map[string]map[string]bool),
		onFire:     onFire,
	}
}

type testSettingsT struct {
	userID       string
	thresholdPct float64
}

func testSettings(userID string, threshold float64) *db.UserSettings {
	return &db.UserSettings{UserID: userID, ThresholdPct: threshold, InAppAlerts: true}
}

func (e *Engine) evaluateForUser(settings *db.UserSettings, opps []aggregator.SpreadOpportunity) {
	uid := settings.UserID
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.activeOpps[uid] == nil {
		e.activeOpps[uid] = make(map[string]bool)
	}
	current := make(map[string]bool)
	for _, opp := range opps {
		if opp.SpreadPct < settings.ThresholdPct {
			continue
		}
		key := oppKey(opp)
		current[key] = true
		if !e.activeOpps[uid][key] {
			e.fireAlert(context.Background(), settings, opp)
		}
	}
	e.activeOpps[uid] = current
}

func (e *Engine) Evaluate(ctx context.Context, opps []aggregator.SpreadOpportunity) {
	if e.queries == nil {
		return
	}
	users, err := e.queries.GetAllSettings(ctx)
	if err != nil {
		return
	}
	for _, s := range users {
		e.evaluateForUser(s, opps)
	}
}

type alertMsg struct {
	Type        string                     `json:"type"`
	Opportunity aggregator.SpreadOpportunity `json:"opportunity"`
}

func (e *Engine) fireAlert(ctx context.Context, settings *db.UserSettings, opp aggregator.SpreadOpportunity) {
	if e.onFire != nil {
		e.onFire(settings.UserID, opp)
		return
	}
	if settings.InAppAlerts && e.hub != nil {
		msg, _ := json.Marshal(alertMsg{Type: "alert", Opportunity: opp})
		e.hub.SendToUser(settings.UserID, msg)
	}
	if settings.TelegramBotToken != "" && settings.TelegramChatID != "" {
		go SendTelegram(settings.TelegramBotToken, settings.TelegramChatID, opp)
	}
	if e.queries != nil {
		_ = e.queries.InsertAlert(ctx, &db.AlertRecord{
			UserID:       settings.UserID,
			Symbol:       opp.Symbol,
			BuyExchange:  opp.BuyAt,
			SellExchange: opp.SellAt,
			SpreadPct:    opp.SpreadPct,
			BuyPrice:     opp.BuyPrice,
			SellPrice:    opp.SellPrice,
			FiredAt:      time.Now(),
		})
	}
}
```

- [ ] **Step 4: Write telegram.go**

```go
package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cryptotrade/app/internal/aggregator"
)

func SendTelegram(botToken, chatID string, opp aggregator.SpreadOpportunity) error {
	text := fmt.Sprintf(
		"Arbitrage Opportunity!\n\nPair: %s\nBuy on %s at $%.4f\nSell on %s at $%.4f\nSpread: +%.4f%%",
		opp.Symbol, opp.BuyAt, opp.BuyPrice, opp.SellAt, opp.SellPrice, opp.SpreadPct,
	)
	payload, _ := json.Marshal(map[string]string{"chat_id": chatID, "text": text})
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
```

- [ ] **Step 5: Run alert tests**

```bash
cd backend && go test ./internal/alert/... -v
```
Expected: all `PASS`

- [ ] **Step 6: Commit**

```bash
git add backend/internal/alert/
git commit -m "feat: add alert engine with debounce and Telegram delivery"
```

---

## Task 13: Auth API + JWT middleware

**Files:**
- Create: `backend/internal/api/auth.go`
- Create: `backend/internal/api/auth_test.go`

- [ ] **Step 1: Write auth_test.go**

```go
package api

import (
	"testing"
	"time"
)

func TestJWT_roundtrip(t *testing.T) {
	secret := "testsecret"
	token, err := generateJWT("user-123", secret)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := validateJWT(token, secret)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != "user-123" {
		t.Fatalf("userID: got %s", claims.UserID)
	}
	if claims.ExpiresAt.Before(time.Now()) {
		t.Fatal("token should not be expired")
	}
}

func TestJWT_wrongSecret(t *testing.T) {
	token, _ := generateJWT("user-123", "secret-a")
	_, err := validateJWT(token, "secret-b")
	if err == nil {
		t.Fatal("expected error with wrong secret")
	}
}
```

- [ ] **Step 2: Run test to confirm failure**

```bash
cd backend && go test ./internal/api/... -run TestJWT -v
```
Expected: compile error.

- [ ] **Step 3: Write auth.go**

```go
package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/cryptotrade/app/internal/db"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func generateJWT(userID, secret string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func validateJWT(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

type ctxKey string

const userIDKey ctxKey = "userID"

func JWTMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenStr string
			if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
				tokenStr = strings.TrimPrefix(auth, "Bearer ")
			} else {
				tokenStr = r.URL.Query().Get("token")
			}
			if tokenStr == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			claims, err := validateJWT(tokenStr, secret)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDKey, claims.UserID)))
		})
	}
}

func userIDFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(userIDKey).(string)
	return v
}

type AuthHandler struct {
	queries   *db.Queries
	jwtSecret string
}

func NewAuthHandler(q *db.Queries, secret string) *AuthHandler {
	return &AuthHandler{queries: q, jwtSecret: secret}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" || len(body.Password) < 8 {
		http.Error(w, "email required; password min 8 chars", http.StatusBadRequest)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	user, err := h.queries.CreateUser(r.Context(), body.Email, string(hash))
	if err != nil {
		http.Error(w, "email already registered", http.StatusConflict)
		return
	}
	token, err := generateJWT(user.ID, h.jwtSecret)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token, "user_id": user.ID})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	user, err := h.queries.GetUserByEmail(r.Context(), body.Email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)) != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	token, err := generateJWT(user.ID, h.jwtSecret)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token, "user_id": user.ID})
}
```

- [ ] **Step 4: Run JWT tests**

```bash
cd backend && go test ./internal/api/... -run TestJWT -v
```
Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add backend/internal/api/auth.go backend/internal/api/auth_test.go
git commit -m "feat: add auth handlers and JWT middleware"
```

---

## Task 14: Settings + History API + WebSocket handler

**Files:**
- Create: `backend/internal/api/settings.go`
- Create: `backend/internal/api/history.go`
- Create: `backend/internal/api/ws.go`

- [ ] **Step 1: Write settings.go**

```go
package api

import (
	"encoding/json"
	"net/http"

	"github.com/cryptotrade/app/internal/db"
)

type SettingsHandler struct{ queries *db.Queries }

func NewSettingsHandler(q *db.Queries) *SettingsHandler { return &SettingsHandler{q} }

func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid := userIDFromCtx(r.Context())
	s, err := h.queries.GetSettings(r.Context(), uid)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func (h *SettingsHandler) Put(w http.ResponseWriter, r *http.Request) {
	uid := userIDFromCtx(r.Context())
	var s db.UserSettings
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	s.UserID = uid
	if err := h.queries.UpsertSettings(r.Context(), &s); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 2: Write history.go**

```go
package api

import (
	"encoding/json"
	"net/http"

	"github.com/cryptotrade/app/internal/db"
)

type HistoryHandler struct{ queries *db.Queries }

func NewHistoryHandler(q *db.Queries) *HistoryHandler { return &HistoryHandler{q} }

func (h *HistoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid := userIDFromCtx(r.Context())
	records, err := h.queries.GetAlertHistory(r.Context(), uid, 100)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if records == nil {
		records = []*db.AlertRecord{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}
```

- [ ] **Step 3: Write ws.go**

```go
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/cryptotrade/app/internal/aggregator"
	"github.com/cryptotrade/app/internal/exchange"
	"github.com/cryptotrade/app/internal/hub"
)

type WSHandler struct {
	hub       *hub.Hub
	agg       *aggregator.Aggregator
	upgrader  *websocket.Upgrader
	jwtSecret string
}

func NewWSHandler(h *hub.Hub, agg *aggregator.Aggregator, upgrader *websocket.Upgrader, secret string) *WSHandler {
	return &WSHandler{hub: h, agg: agg, upgrader: upgrader, jwtSecret: secret}
}

type snapshotMsg struct {
	Type          string                          `json:"type"`
	Ticks         map[string]exchange.PriceTick   `json:"ticks"`
	Opportunities []aggregator.SpreadOpportunity  `json:"opportunities"`
}

func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	claims, err := validateJWT(tokenStr, h.jwtSecret)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &hub.Client{
		UserID: claims.UserID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}
	h.hub.Register(client)
	defer func() {
		h.hub.Unregister(client.UserID)
		conn.Close()
	}()

	// send current state immediately
	snap := snapshotMsg{
		Type:          "snapshot",
		Ticks:         h.agg.GetTicks(),
		Opportunities: h.agg.ComputeOpportunities(),
	}
	if data, err := json.Marshal(snap); err == nil {
		client.Send <- data
	}

	go h.writePump(client)
	h.readPump(client)
}

func (h *WSHandler) writePump(c *hub.Client) {
	ping := time.NewTicker(30 * time.Second)
	defer ping.Stop()
	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, msg)
		case <-ping.C:
			c.Conn.WriteMessage(websocket.PingMessage, nil)
		}
	}
}

func (h *WSHandler) readPump(c *hub.Client) {
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		if _, _, err := c.Conn.ReadMessage(); err != nil {
			break
		}
	}
}
```

- [ ] **Step 4: Verify all compiles**

```bash
cd backend && go build ./...
```
Expected: no output.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/api/
git commit -m "feat: add settings, history, and WebSocket handlers"
```

---

## Task 15: Wire main.go

**Files:**
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Write main.go**

```go
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"

	"github.com/cryptotrade/app/internal/aggregator"
	"github.com/cryptotrade/app/internal/alert"
	"github.com/cryptotrade/app/internal/api"
	"github.com/cryptotrade/app/internal/config"
	"github.com/cryptotrade/app/internal/db"
	"github.com/cryptotrade/app/internal/exchange"
	"github.com/cryptotrade/app/internal/hub"
)

func main() {
	cfg := config.Load()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	if err := db.Migrate(ctx, pool, "migrations"); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	queries := db.NewQueries(pool)
	h := hub.New()
	agg := aggregator.New()
	engine := alert.New(queries, h)

	tickCh := make(chan exchange.PriceTick, 2000)
	go (&exchange.Binance{}).Connect(ctx, tickCh)
	go (&exchange.Coinbase{}).Connect(ctx, tickCh)
	go (&exchange.Kraken{}).Connect(ctx, tickCh)

	go func() {
		broadcast := time.NewTicker(500 * time.Millisecond)
		defer broadcast.Stop()
		for {
			select {
			case tick := <-tickCh:
				agg.UpdateTick(tick)
			case <-broadcast.C:
				opps := agg.ComputeOpportunities()
				engine.Evaluate(ctx, opps)
				snap := map[string]interface{}{
					"type":          "snapshot",
					"ticks":         agg.GetTicks(),
					"opportunities": opps,
				}
				data, _ := json.Marshal(snap)
				h.Broadcast(data)
			case <-ctx.Done():
				return
			}
		}
	}()

	upgrader := &websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	authH := api.NewAuthHandler(queries, cfg.JWTSecret)
	settingsH := api.NewSettingsHandler(queries)
	historyH := api.NewHistoryHandler(queries)
	wsH := api.NewWSHandler(h, agg, upgrader, cfg.JWTSecret)

	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	r.Post("/api/auth/register", authH.Register)
	r.Post("/api/auth/login", authH.Login)
	r.Group(func(r chi.Router) {
		r.Use(api.JWTMiddleware(cfg.JWTSecret))
		r.Get("/api/settings", settingsH.Get)
		r.Put("/api/settings", settingsH.Put)
		r.Get("/api/history", historyH.Get)
	})
	r.Get("/ws", wsH.ServeWS)

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}
	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()
	log.Printf("listening on :%s", cfg.Port)
	srv.ListenAndServe()
}
```

Note: `Binance`, `Coinbase`, `Kraken` structs need to be exported. Update exchange files: change `type binance struct{}` → `type Binance struct{}` (capital B), same for Coinbase and Kraken.

- [ ] **Step 2: Export exchange structs (Binance, Coinbase, Kraken)**

In `binance.go`: change `type Binance struct{}` — it's already exported above. Verify `coinbase.go` and `kraken.go` also use exported names.

- [ ] **Step 3: Build**

```bash
cd backend && go build ./...
```
Expected: no output.

- [ ] **Step 4: Run all tests**

```bash
cd backend && go test ./... -v
```
Expected: all `PASS`

- [ ] **Step 5: Commit**

```bash
git add backend/cmd/server/main.go
git commit -m "feat: wire all backend components in main.go"
```

---

## Task 16: Frontend scaffold

**Files:**
- Create: `frontend/` (Vite + Vue project)

- [ ] **Step 1: Scaffold with Vite**

```bash
cd /Users/mannemsrinivas/projects/cryptotrade
npm create vite@latest frontend -- --template vue
```

- [ ] **Step 2: Install dependencies**

```bash
cd frontend
npm install
npm install pinia vue-router@4
npm install -D tailwindcss@3 postcss autoprefixer
npx tailwindcss init -p
```

- [ ] **Step 3: Configure Tailwind**

`frontend/tailwind.config.js`:
```js
/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js}'],
  theme: { extend: {} },
  plugins: [],
}
```

- [ ] **Step 4: Add Tailwind to CSS**

`frontend/src/style.css`:
```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

- [ ] **Step 5: Configure Vite proxy**

`frontend/vite.config.js`:
```js
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/ws': { target: 'ws://localhost:8080', ws: true },
    },
  },
})
```

- [ ] **Step 6: Write main.js**

`frontend/src/main.js`:
```js
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router/index.js'
import App from './App.vue'
import './style.css'

createApp(App).use(createPinia()).use(router).mount('#app')
```

- [ ] **Step 7: Write router/index.js**

```js
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth.js'

const routes = [
  { path: '/login', component: () => import('@/views/LoginView.vue') },
  { path: '/dashboard', component: () => import('@/views/DashboardView.vue'), meta: { requiresAuth: true } },
  { path: '/prices', component: () => import('@/views/AllPricesView.vue'), meta: { requiresAuth: true } },
  { path: '/history', component: () => import('@/views/AlertHistoryView.vue'), meta: { requiresAuth: true } },
  { path: '/settings', component: () => import('@/views/SettingsView.vue'), meta: { requiresAuth: true } },
  { path: '/', redirect: '/dashboard' },
]

const router = createRouter({ history: createWebHistory(), routes })

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isAuthenticated) return '/login'
})

export default router
```

- [ ] **Step 8: Write stub App.vue**

```vue
<template>
  <NavBar v-if="auth.isAuthenticated" />
  <router-view />
  <AlertToast />
</template>

<script setup>
import { useAuthStore } from '@/stores/auth.js'
import NavBar from '@/components/NavBar.vue'
import AlertToast from '@/components/AlertToast.vue'
const auth = useAuthStore()
</script>
```

- [ ] **Step 9: Verify dev server starts**

```bash
cd frontend && npm run dev
```
Expected: `VITE ready on http://localhost:5173` (will show errors for missing components — that's ok).

- [ ] **Step 10: Commit**

```bash
cd /Users/mannemsrinivas/projects/cryptotrade
git add frontend/
git commit -m "feat: scaffold Vue 3 + Vite + Tailwind + Pinia + Router frontend"
```

---

## Task 17: Auth store + LoginView

**Files:**
- Create: `frontend/src/stores/auth.js`
- Create: `frontend/src/views/LoginView.vue`

- [ ] **Step 1: Write stores/auth.js**

```js
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import router from '@/router/index.js'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userID = ref(localStorage.getItem('user_id') || '')
  const settings = ref(null)

  const isAuthenticated = computed(() => !!token.value)

  async function _setSession(data) {
    token.value = data.token
    userID.value = data.user_id
    localStorage.setItem('token', data.token)
    localStorage.setItem('user_id', data.user_id)
    await fetchSettings()
  }

  async function login(email, password) {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
    if (!res.ok) throw new Error('Invalid credentials')
    await _setSession(await res.json())
    router.push('/dashboard')
  }

  async function register(email, password) {
    const res = await fetch('/api/auth/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
    if (!res.ok) throw new Error((await res.text()) || 'Registration failed')
    await _setSession(await res.json())
    router.push('/dashboard')
  }

  async function fetchSettings() {
    const res = await fetch('/api/settings', {
      headers: { Authorization: `Bearer ${token.value}` },
    })
    if (res.ok) settings.value = await res.json()
  }

  async function updateSettings(s) {
    const res = await fetch('/api/settings', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token.value}` },
      body: JSON.stringify(s),
    })
    if (!res.ok) throw new Error('Failed to save')
    settings.value = { ...settings.value, ...s }
  }

  function logout() {
    token.value = ''
    userID.value = ''
    settings.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user_id')
    router.push('/login')
  }

  return { token, userID, settings, isAuthenticated, login, register, logout, fetchSettings, updateSettings }
})
```

- [ ] **Step 2: Write LoginView.vue**

```vue
<template>
  <div class="min-h-screen bg-gray-50 flex items-center justify-center">
    <div class="bg-white rounded-xl shadow p-8 w-full max-w-sm">
      <h1 class="text-2xl font-bold text-gray-900 mb-6 text-center">CryptoTrade</h1>

      <div class="flex rounded-lg border border-gray-200 mb-6">
        <button
          v-for="tab in ['Login', 'Register']" :key="tab"
          @click="mode = tab"
          class="flex-1 py-2 text-sm font-medium rounded-lg transition"
          :class="mode === tab ? 'bg-indigo-600 text-white' : 'text-gray-600 hover:bg-gray-50'"
        >{{ tab }}</button>
      </div>

      <form @submit.prevent="submit" class="space-y-4">
        <input
          v-model="email" type="email" placeholder="Email" required
          class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
        <input
          v-model="password" type="password" placeholder="Password (min 8 chars)" required
          class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
        <p v-if="error" class="text-red-500 text-sm">{{ error }}</p>
        <button
          type="submit" :disabled="loading"
          class="w-full bg-indigo-600 text-white py-2 rounded-lg text-sm font-medium hover:bg-indigo-700 disabled:opacity-50"
        >{{ loading ? 'Please wait...' : mode }}</button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth.js'

const auth = useAuthStore()
const mode = ref('Login')
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    if (mode.value === 'Login') await auth.login(email.value, password.value)
    else await auth.register(email.value, password.value)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/stores/auth.js frontend/src/views/LoginView.vue
git commit -m "feat: add auth store and login/register view"
```

---

## Task 18: Prices store + WebSocket composable

**Files:**
- Create: `frontend/src/stores/prices.js`
- Create: `frontend/src/composables/useWebSocket.js`

- [ ] **Step 1: Write stores/prices.js**

```js
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAuthStore } from './auth.js'

export const usePricesStore = defineStore('prices', () => {
  const ticks = ref(new Map())   // "exchange:symbol" → PriceTick
  const alertQueue = ref([])

  const opportunities = computed(() => {
    const auth = useAuthStore()
    const threshold = auth.settings?.threshold_pct ?? 0.1

    const bySymbol = new Map()
    for (const [, tick] of ticks.value) {
      if (!bySymbol.has(tick.symbol)) bySymbol.set(tick.symbol, [])
      bySymbol.get(tick.symbol).push(tick)
    }

    const opps = []
    for (const [symbol, ts] of bySymbol) {
      for (let i = 0; i < ts.length; i++) {
        for (let j = 0; j < ts.length; j++) {
          if (i === j) continue
          const buyPrice = ts[i].ask
          const sellPrice = ts[j].bid
          if (!buyPrice || !sellPrice) continue
          const spreadPct = (sellPrice - buyPrice) / buyPrice * 100
          if (spreadPct >= threshold) {
            opps.push({ symbol, buyAt: ts[i].exchange, sellAt: ts[j].exchange, buyPrice, sellPrice, spreadPct })
          }
        }
      }
    }
    return opps.sort((a, b) => b.spreadPct - a.spreadPct)
  })

  function applySnapshot(snapshot) {
    const next = new Map()
    for (const [key, tick] of Object.entries(snapshot.ticks || {})) {
      next.set(key, tick)
    }
    ticks.value = next
  }

  function pushAlert(opp) { alertQueue.value.push(opp) }
  function shiftAlert() { return alertQueue.value.shift() }

  return { ticks, alertQueue, opportunities, applySnapshot, pushAlert, shiftAlert }
})
```

- [ ] **Step 2: Write composables/useWebSocket.js**

```js
import { useAuthStore } from '@/stores/auth.js'
import { usePricesStore } from '@/stores/prices.js'

let ws = null
let reconnectTimer = null
let backoff = 1000

export function useWebSocket() {
  function connect() {
    const auth = useAuthStore()
    if (!auth.token) return

    ws = new WebSocket(`/ws?token=${auth.token}`)

    ws.onopen = () => { backoff = 1000 }

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data)
      const prices = usePricesStore()
      if (msg.type === 'snapshot') {
        prices.applySnapshot(msg)
      } else if (msg.type === 'alert') {
        prices.pushAlert(msg.opportunity)
      }
    }

    ws.onclose = () => { scheduleReconnect() }
    ws.onerror = () => { ws?.close() }
  }

  function scheduleReconnect() {
    if (reconnectTimer) return
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      backoff = Math.min(backoff * 2, 30000)
      connect()
    }, backoff)
  }

  function disconnect() {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
    ws?.close()
    ws = null
  }

  return { connect, disconnect }
}
```

- [ ] **Step 3: Connect WebSocket on login — update App.vue**

```vue
<template>
  <NavBar v-if="auth.isAuthenticated" />
  <router-view />
  <AlertToast />
</template>

<script setup>
import { watch } from 'vue'
import { useAuthStore } from '@/stores/auth.js'
import { useWebSocket } from '@/composables/useWebSocket.js'
import NavBar from '@/components/NavBar.vue'
import AlertToast from '@/components/AlertToast.vue'

const auth = useAuthStore()
const { connect, disconnect } = useWebSocket()

watch(() => auth.isAuthenticated, (authed) => {
  if (authed) connect()
  else disconnect()
}, { immediate: true })
</script>
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/stores/prices.js frontend/src/composables/useWebSocket.js frontend/src/App.vue
git commit -m "feat: add prices store and WebSocket composable with auto-reconnect"
```

---

## Task 19: NavBar + StatCard + OpportunityTable + DashboardView

**Files:**
- Create: `frontend/src/components/NavBar.vue`
- Create: `frontend/src/components/StatCard.vue`
- Create: `frontend/src/components/OpportunityTable.vue`
- Create: `frontend/src/views/DashboardView.vue`

- [ ] **Step 1: Write NavBar.vue**

```vue
<template>
  <nav class="bg-white border-b border-gray-200 px-6 py-3 flex items-center justify-between">
    <span class="font-bold text-indigo-600 text-lg">CryptoTrade</span>
    <div class="flex gap-6 text-sm font-medium">
      <RouterLink to="/dashboard" class="text-gray-600 hover:text-indigo-600" active-class="text-indigo-600">Dashboard</RouterLink>
      <RouterLink to="/prices" class="text-gray-600 hover:text-indigo-600" active-class="text-indigo-600">All Prices</RouterLink>
      <RouterLink to="/history" class="text-gray-600 hover:text-indigo-600" active-class="text-indigo-600">Alert History</RouterLink>
      <RouterLink to="/settings" class="text-gray-600 hover:text-indigo-600" active-class="text-indigo-600">Settings</RouterLink>
    </div>
    <button @click="auth.logout()" class="text-sm text-gray-500 hover:text-red-500">Logout</button>
  </nav>
</template>

<script setup>
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth.js'
const auth = useAuthStore()
</script>
```

- [ ] **Step 2: Write StatCard.vue**

```vue
<template>
  <div class="bg-white rounded-xl border border-gray-200 p-5">
    <p class="text-xs font-semibold text-gray-500 uppercase tracking-wide">{{ label }}</p>
    <p class="text-3xl font-bold mt-1" :class="valueClass">{{ value }}</p>
    <p v-if="sub" class="text-xs text-gray-400 mt-1">{{ sub }}</p>
  </div>
</template>

<script setup>
defineProps({
  label: String,
  value: [String, Number],
  sub: String,
  valueClass: { type: String, default: 'text-gray-900' },
})
</script>
```

- [ ] **Step 3: Write OpportunityTable.vue**

```vue
<template>
  <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
    <table class="w-full text-sm">
      <thead class="bg-gray-50 border-b border-gray-200">
        <tr>
          <th class="text-left px-4 py-3 font-semibold text-gray-600">Pair</th>
          <th class="text-left px-4 py-3 font-semibold text-gray-600">Buy At</th>
          <th class="text-left px-4 py-3 font-semibold text-gray-600">Buy Price</th>
          <th class="text-left px-4 py-3 font-semibold text-gray-600">Sell At</th>
          <th class="text-left px-4 py-3 font-semibold text-gray-600">Sell Price</th>
          <th class="text-right px-4 py-3 font-semibold text-gray-600">Spread %</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="opportunities.length === 0">
          <td colspan="6" class="text-center py-12 text-gray-400">No opportunities above your threshold right now.</td>
        </tr>
        <tr
          v-for="opp in opportunities" :key="opp.symbol + opp.buyAt + opp.sellAt"
          class="border-t border-gray-100 bg-green-50 border-l-4 border-l-green-400"
        >
          <td class="px-4 py-3 font-medium text-gray-900">{{ opp.symbol }}</td>
          <td class="px-4 py-3 capitalize text-gray-700">{{ opp.buyAt }}</td>
          <td class="px-4 py-3 text-gray-700">${{ opp.buyPrice.toFixed(4) }}</td>
          <td class="px-4 py-3 capitalize text-gray-700">{{ opp.sellAt }}</td>
          <td class="px-4 py-3 text-gray-700">${{ opp.sellPrice.toFixed(4) }}</td>
          <td class="px-4 py-3 text-right font-bold text-green-600">+{{ opp.spreadPct.toFixed(4) }}%</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { usePricesStore } from '@/stores/prices.js'
import { storeToRefs } from 'pinia'
const { opportunities } = storeToRefs(usePricesStore())
</script>
```

- [ ] **Step 4: Write DashboardView.vue**

```vue
<template>
  <div class="min-h-screen bg-gray-50 p-6">
    <h2 class="text-xl font-bold text-gray-900 mb-5">Dashboard</h2>

    <div class="grid grid-cols-3 gap-4 mb-6">
      <StatCard
        label="Active Opportunities"
        :value="opportunities.length"
        value-class="text-green-600"
        :sub="`threshold: ${threshold}%`"
      />
      <StatCard
        label="Avg Spread"
        :value="avgSpread ? avgSpread + '%' : '—'"
        value-class="text-gray-900"
      />
      <StatCard
        label="Tracked Pairs"
        :value="trackedPairs"
        value-class="text-gray-900"
      />
    </div>

    <OpportunityTable />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { usePricesStore } from '@/stores/prices.js'
import { useAuthStore } from '@/stores/auth.js'
import { storeToRefs } from 'pinia'
import StatCard from '@/components/StatCard.vue'
import OpportunityTable from '@/components/OpportunityTable.vue'

const { opportunities, ticks } = storeToRefs(usePricesStore())
const auth = useAuthStore()
const threshold = computed(() => auth.settings?.threshold_pct ?? 0.1)
const avgSpread = computed(() => {
  if (!opportunities.value.length) return null
  const avg = opportunities.value.reduce((s, o) => s + o.spreadPct, 0) / opportunities.value.length
  return avg.toFixed(4)
})
const trackedPairs = computed(() => new Set([...ticks.value.values()].map(t => t.symbol)).size)
</script>
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/NavBar.vue frontend/src/components/StatCard.vue \
        frontend/src/components/OpportunityTable.vue frontend/src/views/DashboardView.vue
git commit -m "feat: add NavBar, StatCard, OpportunityTable, and Dashboard view"
```

---

## Task 20: PriceGrid + AllPricesView

**Files:**
- Create: `frontend/src/components/PriceGrid.vue`
- Create: `frontend/src/views/AllPricesView.vue`

- [ ] **Step 1: Write PriceGrid.vue**

```vue
<template>
  <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
    <table class="w-full text-sm">
      <thead class="bg-gray-50 border-b border-gray-200">
        <tr>
          <th class="text-left px-4 py-3 font-semibold text-gray-600">Pair</th>
          <th class="text-left px-4 py-3 font-semibold text-gray-600">Binance</th>
          <th class="text-left px-4 py-3 font-semibold text-gray-600">Coinbase</th>
          <th class="text-left px-4 py-3 font-semibold text-gray-600">Kraken</th>
          <th
            class="text-right px-4 py-3 font-semibold text-gray-600 cursor-pointer select-none"
            @click="$emit('sort')"
          >Best Spread % ↕</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in rows" :key="row.symbol" class="border-t border-gray-100 hover:bg-gray-50">
          <td class="px-4 py-3 font-medium text-gray-900">{{ row.symbol }}</td>
          <td class="px-4 py-3 text-gray-700">{{ fmt(row.binance) }}</td>
          <td class="px-4 py-3 text-gray-700">{{ fmt(row.coinbase) }}</td>
          <td class="px-4 py-3 text-gray-700">{{ fmt(row.kraken) }}</td>
          <td class="px-4 py-3 text-right font-medium" :class="row.bestSpread > 0 ? 'text-green-600' : 'text-gray-400'">
            {{ row.bestSpread > 0 ? '+' + row.bestSpread.toFixed(4) + '%' : '—' }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { usePricesStore } from '@/stores/prices.js'
import { storeToRefs } from 'pinia'

const props = defineProps({ sortDesc: { type: Boolean, default: true } })
defineEmits(['sort'])

const { ticks } = storeToRefs(usePricesStore())

const EXCHANGES = ['binance', 'coinbase', 'kraken']

const rows = computed(() => {
  const bySymbol = new Map()
  for (const [, tick] of ticks.value) {
    if (!bySymbol.has(tick.symbol)) bySymbol.set(tick.symbol, {})
    bySymbol.get(tick.symbol)[tick.exchange] = tick
  }
  const result = []
  for (const [symbol, exMap] of bySymbol) {
    const ts = EXCHANGES.map(e => exMap[e]).filter(Boolean)
    let bestSpread = 0
    for (let i = 0; i < ts.length; i++) {
      for (let j = 0; j < ts.length; j++) {
        if (i === j) continue
        const sp = (ts[j].bid - ts[i].ask) / ts[i].ask * 100
        if (sp > bestSpread) bestSpread = sp
      }
    }
    result.push({ symbol, binance: exMap.binance, coinbase: exMap.coinbase, kraken: exMap.kraken, bestSpread })
  }
  return result.sort((a, b) => props.sortDesc ? b.bestSpread - a.bestSpread : a.bestSpread - b.bestSpread)
})

function fmt(tick) {
  if (!tick) return '—'
  return `$${tick.bid.toFixed(2)} / $${tick.ask.toFixed(2)}`
}
</script>
```

- [ ] **Step 2: Write AllPricesView.vue**

```vue
<template>
  <div class="min-h-screen bg-gray-50 p-6">
    <div class="flex items-center justify-between mb-5">
      <h2 class="text-xl font-bold text-gray-900">All Prices</h2>
      <span class="text-xs text-gray-400">Updates every 500ms</span>
    </div>
    <PriceGrid :sort-desc="sortDesc" @sort="sortDesc = !sortDesc" />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import PriceGrid from '@/components/PriceGrid.vue'
const sortDesc = ref(true)
</script>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/PriceGrid.vue frontend/src/views/AllPricesView.vue
git commit -m "feat: add PriceGrid and All Prices view"
```

---

## Task 21: AlertToast + AlertHistoryView + SettingsView

**Files:**
- Create: `frontend/src/components/AlertToast.vue`
- Create: `frontend/src/views/AlertHistoryView.vue`
- Create: `frontend/src/views/SettingsView.vue`

- [ ] **Step 1: Write AlertToast.vue**

```vue
<template>
  <Teleport to="body">
    <div class="fixed bottom-5 right-5 space-y-2 z-50">
      <TransitionGroup name="toast">
        <div
          v-for="toast in toasts" :key="toast.id"
          class="bg-white border border-green-300 shadow-lg rounded-xl p-4 w-80"
        >
          <p class="font-semibold text-green-700 text-sm">Opportunity: {{ toast.symbol }}</p>
          <p class="text-xs text-gray-600 mt-1">
            Buy {{ toast.buyAt }} ${{ toast.buyPrice.toFixed(4) }} →
            Sell {{ toast.sellAt }} ${{ toast.sellPrice.toFixed(4) }}
          </p>
          <p class="text-xs font-bold text-green-600 mt-1">+{{ toast.spreadPct.toFixed(4) }}%</p>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, watch } from 'vue'
import { usePricesStore } from '@/stores/prices.js'
import { useAuthStore } from '@/stores/auth.js'

const prices = usePricesStore()
const toasts = ref([])
let id = 0

watch(() => prices.alertQueue.length, () => {
  const opp = prices.shiftAlert()
  if (!opp) return
  const toast = { ...opp, id: ++id }
  toasts.value.push(toast)
  const auth = useAuthStore()
  if (auth.settings?.alert_sound) {
    // Short 440Hz beep via Web Audio API — no external asset needed
    const ctx = new AudioContext()
    const osc = ctx.createOscillator()
    osc.connect(ctx.destination)
    osc.frequency.value = 440
    osc.start()
    osc.stop(ctx.currentTime + 0.15)
  }
  setTimeout(() => {
    toasts.value = toasts.value.filter(t => t.id !== toast.id)
  }, 6000)
})
</script>

<style scoped>
.toast-enter-active, .toast-leave-active { transition: all 0.3s ease; }
.toast-enter-from, .toast-leave-to { opacity: 0; transform: translateX(40px); }
</style>
```

- [ ] **Step 2: Write AlertHistoryView.vue**

```vue
<template>
  <div class="min-h-screen bg-gray-50 p-6">
    <h2 class="text-xl font-bold text-gray-900 mb-5">Alert History</h2>
    <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
      <table class="w-full text-sm">
        <thead class="bg-gray-50 border-b border-gray-200">
          <tr>
            <th class="text-left px-4 py-3 font-semibold text-gray-600">Time</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600">Pair</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600">Buy</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600">Sell</th>
            <th class="text-right px-4 py-3 font-semibold text-gray-600">Spread %</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="!records.length">
            <td colspan="5" class="text-center py-12 text-gray-400">No alerts yet.</td>
          </tr>
          <tr v-for="r in records" :key="r.id" class="border-t border-gray-100 hover:bg-gray-50">
            <td class="px-4 py-3 text-gray-500">{{ fmtDate(r.fired_at) }}</td>
            <td class="px-4 py-3 font-medium text-gray-900">{{ r.symbol }}</td>
            <td class="px-4 py-3 capitalize text-gray-700">{{ r.buy_exchange }} ${{ r.buy_price.toFixed(4) }}</td>
            <td class="px-4 py-3 capitalize text-gray-700">{{ r.sell_exchange }} ${{ r.sell_price.toFixed(4) }}</td>
            <td class="px-4 py-3 text-right font-bold text-green-600">+{{ r.spread_pct.toFixed(4) }}%</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth.js'

const auth = useAuthStore()
const records = ref([])

onMounted(async () => {
  const res = await fetch('/api/history', { headers: { Authorization: `Bearer ${auth.token}` } })
  if (res.ok) records.value = await res.json()
})

function fmtDate(iso) {
  return new Date(iso).toLocaleString()
}
</script>
```

- [ ] **Step 3: Write SettingsView.vue**

```vue
<template>
  <div class="min-h-screen bg-gray-50 p-6">
    <h2 class="text-xl font-bold text-gray-900 mb-5">Settings</h2>
    <div class="bg-white rounded-xl border border-gray-200 p-6 max-w-lg">
      <form @submit.prevent="save" class="space-y-5">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Alert Threshold (%)</label>
          <input v-model.number="form.threshold_pct" type="number" min="0" max="100" step="0.01"
            class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
          <p class="text-xs text-gray-400 mt-1">Minimum spread % to trigger an alert</p>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Telegram Bot Token</label>
          <input v-model="form.telegram_bot_token" type="text" placeholder="123456:ABC..."
            class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Telegram Chat ID</label>
          <input v-model="form.telegram_chat_id" type="text" placeholder="-1001234567890"
            class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
        </div>
        <div class="flex items-center justify-between">
          <label class="text-sm font-medium text-gray-700">In-app alerts</label>
          <button type="button" @click="form.in_app_alerts = !form.in_app_alerts"
            :class="form.in_app_alerts ? 'bg-indigo-600' : 'bg-gray-200'"
            class="relative w-10 h-6 rounded-full transition-colors">
            <span :class="form.in_app_alerts ? 'translate-x-4' : 'translate-x-1'"
              class="absolute top-1 w-4 h-4 bg-white rounded-full shadow transition-transform"></span>
          </button>
        </div>
        <div class="flex items-center justify-between">
          <label class="text-sm font-medium text-gray-700">Alert sound</label>
          <button type="button" @click="form.alert_sound = !form.alert_sound"
            :class="form.alert_sound ? 'bg-indigo-600' : 'bg-gray-200'"
            class="relative w-10 h-6 rounded-full transition-colors">
            <span :class="form.alert_sound ? 'translate-x-4' : 'translate-x-1'"
              class="absolute top-1 w-4 h-4 bg-white rounded-full shadow transition-transform"></span>
          </button>
        </div>
        <p v-if="saved" class="text-green-600 text-sm">Settings saved.</p>
        <p v-if="err" class="text-red-500 text-sm">{{ err }}</p>
        <button type="submit"
          class="w-full bg-indigo-600 text-white py-2 rounded-lg text-sm font-medium hover:bg-indigo-700">
          Save Settings
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth.js'

const auth = useAuthStore()
const saved = ref(false)
const err = ref('')
const form = ref({
  threshold_pct: 0.1,
  telegram_bot_token: '',
  telegram_chat_id: '',
  in_app_alerts: true,
  alert_sound: true,
})

onMounted(() => {
  if (auth.settings) Object.assign(form.value, auth.settings)
})

async function save() {
  saved.value = false
  err.value = ''
  try {
    await auth.updateSettings(form.value)
    saved.value = true
    setTimeout(() => { saved.value = false }, 3000)
  } catch (e) {
    err.value = e.message
  }
}
</script>
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/AlertToast.vue \
        frontend/src/views/AlertHistoryView.vue \
        frontend/src/views/SettingsView.vue
git commit -m "feat: add AlertToast, Alert History view, and Settings view"
```

---

## Verification

### Start the stack

```bash
# 1. Ensure PostgreSQL is running with a database named "cryptotrade"
createdb cryptotrade   # or via psql

# 2. Backend
cd backend && go run ./cmd/server

# 3. Frontend (new terminal)
cd frontend && npm run dev
```

Open http://localhost:5173

### End-to-end checklist

- [ ] Register a new user → redirected to Dashboard
- [ ] Dashboard shows "0 active opportunities" stat card with WebSocket connected
- [ ] Navigate to All Prices → prices populate within 5 seconds (live Binance/Coinbase/Kraken feeds)
- [ ] Go to Settings → set threshold to `0.01` → save
- [ ] Return to Dashboard → an opportunity should appear in the table (green row)
- [ ] Wait for alert toast to appear in bottom-right corner
- [ ] Set Telegram bot token + chat ID in Settings → verify message arrives on Telegram
- [ ] Navigate to Alert History → past alerts listed with timestamp and spread %
- [ ] Set threshold back to `0.5` → opportunity table clears to empty state message
- [ ] Open a second browser tab, register a second user with a different threshold → verify each user only receives alerts for their own threshold
