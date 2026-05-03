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

func newTestEngine(onFire func(string, aggregator.SpreadOpportunity)) *Engine {
	return &Engine{
		activeOpps: make(map[string]map[string]bool),
		onFire:     onFire,
	}
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
	Type        string                       `json:"type"`
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
