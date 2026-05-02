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
