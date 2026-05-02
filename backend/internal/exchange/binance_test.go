package exchange

import "testing"

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
