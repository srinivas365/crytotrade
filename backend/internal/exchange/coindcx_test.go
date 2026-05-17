package exchange

import "testing"

func TestCoinDCX_parseTickers_USDT(t *testing.T) {
	raw := `[{"market":"BTCUSDT","bid":"63240.00","ask":"63242.50","last_price":"63241.00"}]`
	c := &CoinDCX{}
	ticks, err := c.parseTickers([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(ticks) != 1 {
		t.Fatalf("ticks len: got %d want 1", len(ticks))
	}
	if ticks[0].Exchange != "coindcx" {
		t.Fatalf("exchange: got %s", ticks[0].Exchange)
	}
	if ticks[0].Symbol != "BTC/USDT" {
		t.Fatalf("symbol: got %s", ticks[0].Symbol)
	}
	if ticks[0].Bid != 63240.00 || ticks[0].Ask != 63242.50 {
		t.Fatalf("bid/ask: got %f/%f", ticks[0].Bid, ticks[0].Ask)
	}
}

func TestCoinDCX_parseTickers_INR(t *testing.T) {
	raw := `[{"market":"BTCINR","bid":"5400000","ask":"5400500","last_price":"5400250"}]`
	c := &CoinDCX{}
	ticks, err := c.parseTickers([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(ticks) != 1 || ticks[0].Symbol != "BTC/INR" {
		t.Fatalf("expected BTC/INR tick, got %+v", ticks)
	}
	if ticks[0].Bid != 5400000 || ticks[0].Ask != 5400500 {
		t.Fatalf("bid/ask: got %f/%f", ticks[0].Bid, ticks[0].Ask)
	}
}

func TestCoinDCX_parseTickers_FallbackToLastPrice(t *testing.T) {
	raw := `[{"market":"ETHUSDT","bid":"0","ask":"0","last_price":"3200.50"}]`
	c := &CoinDCX{}
	ticks, err := c.parseTickers([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(ticks) != 1 {
		t.Fatalf("ticks len: got %d", len(ticks))
	}
	if ticks[0].Bid != 3200.50 || ticks[0].Ask != 3200.50 {
		t.Fatalf("expected last_price fallback for both bid/ask, got %f/%f", ticks[0].Bid, ticks[0].Ask)
	}
}

func TestCoinDCX_parseTickers_UnknownSymbol(t *testing.T) {
	raw := `[{"market":"XYZUSDT","bid":"1.0","ask":"1.1","last_price":"1.05"}]`
	c := &CoinDCX{}
	ticks, err := c.parseTickers([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(ticks) != 0 {
		t.Fatalf("expected empty result for unknown symbol, got %+v", ticks)
	}
}

// Real CoinDCX responses contain at least one ticker (e.g. BTCINR_insta) where
// bid/ask come back as bare JSON numbers instead of quoted strings, with nulls
// for last_price/high/low/volume. Before flexFloat, this single outlier failed
// the whole array decode and broke the connector.
func TestCoinDCX_parseTickers_MixedNumberAndStringTypes(t *testing.T) {
	raw := `[
		{"market":"BTCUSDT","bid":"63240","ask":"63242","last_price":"63241","high":"64000","low":"62000","volume":"100"},
		{"market":"BTCINR_insta","bid":4248058.33,"ask":4275454.77,"last_price":null,"high":null,"low":null,"volume":null}
	]`
	c := &CoinDCX{}
	ticks, err := c.parseTickers([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error decoding mixed payload: %v", err)
	}
	if len(ticks) != 1 || ticks[0].Symbol != "BTC/USDT" {
		t.Fatalf("expected BTC/USDT (BTCINR_insta is not in our map and should be filtered), got %+v", ticks)
	}
}

func TestCoinDCX_parseTickers_MultiplePairs(t *testing.T) {
	raw := `[
		{"market":"BTCUSDT","bid":"63240","ask":"63242","last_price":"63241"},
		{"market":"BTCINR","bid":"5400000","ask":"5400500","last_price":"5400250"},
		{"market":"UNKNOWN","bid":"1","ask":"2","last_price":"1.5"}
	]`
	c := &CoinDCX{}
	ticks, err := c.parseTickers([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(ticks) != 2 {
		t.Fatalf("expected 2 ticks (unknown filtered), got %d", len(ticks))
	}
}
