package exchange

import "testing"

func TestIndepReserve_parseSummary_Standard(t *testing.T) {
	raw := `{
		"DayHighestPrice":101907.23,
		"DayLowestPrice":99109.95,
		"DayAvgPrice":100508.59,
		"CurrentLowestOfferPrice":100225.19,
		"CurrentHighestBidPrice":100062.37,
		"LastPrice":99634.05,
		"PrimaryCurrencyCode":"Xbt",
		"SecondaryCurrencyCode":"Sgd",
		"CreatedTimestampUtc":"2026-05-17T05:29:51.7822783Z"
	}`
	i := &IndepReserve{}
	tick, err := i.parseSummary([]byte(raw), "BTC/SGD")
	if err != nil {
		t.Fatal(err)
	}
	if tick == nil {
		t.Fatal("expected tick, got nil")
	}
	if tick.Exchange != "indep_reserve" {
		t.Fatalf("exchange: got %s", tick.Exchange)
	}
	if tick.Symbol != "BTC/SGD" {
		t.Fatalf("symbol: got %s", tick.Symbol)
	}
	if tick.Bid != 100062.37 || tick.Ask != 100225.19 {
		t.Fatalf("bid/ask: got %f/%f", tick.Bid, tick.Ask)
	}
}

func TestIndepReserve_parseSummary_FallbackToLastPrice(t *testing.T) {
	raw := `{"CurrentHighestBidPrice":0,"CurrentLowestOfferPrice":0,"LastPrice":50.25}`
	i := &IndepReserve{}
	tick, err := i.parseSummary([]byte(raw), "FOO/SGD")
	if err != nil {
		t.Fatal(err)
	}
	if tick == nil {
		t.Fatal("expected tick from last-price fallback")
	}
	if tick.Bid != 50.25 || tick.Ask != 50.25 {
		t.Fatalf("expected last_price fallback for both bid/ask, got %f/%f", tick.Bid, tick.Ask)
	}
}

func TestIndepReserve_parseSummary_NoUsablePrice(t *testing.T) {
	raw := `{"CurrentHighestBidPrice":0,"CurrentLowestOfferPrice":0,"LastPrice":0}`
	i := &IndepReserve{}
	tick, err := i.parseSummary([]byte(raw), "FOO/SGD")
	if err != nil {
		t.Fatal(err)
	}
	if tick != nil {
		t.Fatalf("expected nil tick when no price available, got %+v", tick)
	}
}

func TestIndepReserve_parseSummary_MalformedJSON(t *testing.T) {
	i := &IndepReserve{}
	_, err := i.parseSummary([]byte(`<html>...`), "BTC/SGD")
	if err == nil {
		t.Fatal("expected error on malformed JSON")
	}
}
