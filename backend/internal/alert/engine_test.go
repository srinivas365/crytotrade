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
