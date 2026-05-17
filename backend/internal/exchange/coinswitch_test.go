package exchange

import "testing"

func TestCoinSwitch_parseTickers_Standard(t *testing.T) {
	raw := `{"data":{
		"BTC/INR":{"openPrice":"7751100","baseVolume":"0.934","quoteVolume":"7141498","lowPrice":"7600187","highPrice":"7751100","lastPrice":"7677813","at":1778996438423,"askPrice":"7677813","bidPrice":"7620423","percentageChange":"-0.95","exchange":"coinswitchx","symbol":"BTC/INR"},
		"ETH/INR":{"askPrice":"215190","bidPrice":"213601","lastPrice":"214500","exchange":"coinswitchx","symbol":"ETH/INR"}
	}}`
	c := &CoinSwitch{}
	ticks, err := c.parseTickers([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(ticks) != 2 {
		t.Fatalf("ticks: got %d want 2", len(ticks))
	}
	for _, tk := range ticks {
		if tk.Exchange != "coinswitch" {
			t.Fatalf("exchange: got %s", tk.Exchange)
		}
		switch tk.Symbol {
		case "BTC/INR":
			if tk.Bid != 7620423 || tk.Ask != 7677813 {
				t.Fatalf("BTC/INR bid/ask: got %f/%f", tk.Bid, tk.Ask)
			}
		case "ETH/INR":
			if tk.Bid != 213601 || tk.Ask != 215190 {
				t.Fatalf("ETH/INR bid/ask: got %f/%f", tk.Bid, tk.Ask)
			}
		default:
			t.Fatalf("unexpected symbol: %s", tk.Symbol)
		}
	}
}

func TestCoinSwitch_parseTickers_FallbackToLastPrice(t *testing.T) {
	raw := `{"data":{"SOL/INR":{"bidPrice":"0","askPrice":"0","lastPrice":"8500.50"}}}`
	c := &CoinSwitch{}
	ticks, err := c.parseTickers([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(ticks) != 1 {
		t.Fatalf("ticks: got %d", len(ticks))
	}
	if ticks[0].Bid != 8500.50 || ticks[0].Ask != 8500.50 {
		t.Fatalf("expected last-price fallback, got %f/%f", ticks[0].Bid, ticks[0].Ask)
	}
}

func TestCoinSwitch_parseTickers_UnknownSymbolFiltered(t *testing.T) {
	raw := `{"data":{
		"0G/INR":{"bidPrice":"50","askPrice":"51","lastPrice":"50.5"},
		"BTC/INR":{"bidPrice":"7620000","askPrice":"7680000","lastPrice":"7650000"}
	}}`
	c := &CoinSwitch{}
	ticks, err := c.parseTickers([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(ticks) != 1 || ticks[0].Symbol != "BTC/INR" {
		t.Fatalf("expected only BTC/INR, got %+v", ticks)
	}
}

func TestCoinSwitch_parseTickers_MalformedJSON(t *testing.T) {
	c := &CoinSwitch{}
	_, err := c.parseTickers([]byte(`<html>`))
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}
