package exchange

import "testing"

func TestCoinbase_parseTick(t *testing.T) {
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
