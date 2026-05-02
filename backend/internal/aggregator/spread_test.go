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
	expected := (63310.0 - 63199.0) / 63199.0 * 100
	if absF(opp.SpreadPct-expected) > 0.0001 {
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

func TestAggregator_ComputeOpportunities(t *testing.T) {
	agg := New()
	agg.UpdateTick(makeTick("coinbase", "BTC/USDT", 63198, 63199))
	agg.UpdateTick(makeTick("kraken", "BTC/USDT", 63310, 63311))

	opps := agg.ComputeOpportunities()
	if len(opps) == 0 {
		t.Fatal("expected at least one opportunity")
	}
	found := false
	for _, o := range opps {
		if o.SpreadPct > 0 && o.BuyAt == "coinbase" && o.SellAt == "kraken" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected coinbase→kraken opportunity")
	}
}

func absF(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
