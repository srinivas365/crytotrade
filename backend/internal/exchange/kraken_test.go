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
