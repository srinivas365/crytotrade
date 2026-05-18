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
	go (&exchange.CoinSwitch{}).Connect(ctx, tickCh)
	go (&exchange.IndepReserve{}).Connect(ctx, tickCh)

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
				snap := map[string]any{
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
