package exchange

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// indepReservePairs maps (primary, secondary) currency codes as used by the
// Independent Reserve REST API to our canonical symbol. Only SGD-quoted pairs
// are tracked here — IR also offers AUD/USD/NZD quotes, but the dashboard's
// focus is SGD for this exchange.
var indepReservePairs = []struct {
	Primary   string // e.g. "Xbt"
	Secondary string // e.g. "Sgd"
	Symbol    string // e.g. "BTC/SGD"
}{
	{"Xbt", "Sgd", "BTC/SGD"},
	{"Eth", "Sgd", "ETH/SGD"},
	{"Sol", "Sgd", "SOL/SGD"},
	{"Xrp", "Sgd", "XRP/SGD"},
	{"Ada", "Sgd", "ADA/SGD"},
	{"Doge", "Sgd", "DOGE/SGD"},
	{"Ltc", "Sgd", "LTC/SGD"},
	{"Dot", "Sgd", "DOT/SGD"},
	{"Matic", "Sgd", "MATIC/SGD"},
}

type IndepReserve struct{}

type indepReserveSummary struct {
	CurrentHighestBidPrice  float64 `json:"CurrentHighestBidPrice"`
	CurrentLowestOfferPrice float64 `json:"CurrentLowestOfferPrice"`
	LastPrice               float64 `json:"LastPrice"`
}

func (i *IndepReserve) parseSummary(raw []byte, symbol string) (*PriceTick, error) {
	var s indepReserveSummary
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	bid, ask := s.CurrentHighestBidPrice, s.CurrentLowestOfferPrice
	if bid <= 0 {
		bid = s.LastPrice
	}
	if ask <= 0 {
		ask = s.LastPrice
	}
	if bid <= 0 || ask <= 0 {
		return nil, nil
	}
	return &PriceTick{
		Exchange:  "indep_reserve",
		Symbol:    symbol,
		Bid:       bid,
		Ask:       ask,
		Timestamp: time.Now(),
	}, nil
}

func (i *IndepReserve) fetchPair(ctx context.Context, client *http.Client, primary, secondary string) ([]byte, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   "api.independentreserve.com",
		Path:   "/Public/GetMarketSummary",
	}
	q := u.Query()
	q.Set("primaryCurrencyCode", primary)
	q.Set("secondaryCurrencyCode", secondary)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// Connect spawns one polling goroutine per pair so a slow or failing pair
// can't stall the rest. Each pair polls every 5s with independent retry.
func (i *IndepReserve) Connect(ctx context.Context, out chan<- PriceTick) {
	client := &http.Client{Timeout: 10 * time.Second}
	var wg sync.WaitGroup
	for _, p := range indepReservePairs {
		wg.Add(1)
		go func(primary, secondary, symbol string) {
			defer wg.Done()
			i.pollPair(ctx, client, primary, secondary, symbol, out)
		}(p.Primary, p.Secondary, p.Symbol)
	}
	wg.Wait()
}

func (i *IndepReserve) pollPair(ctx context.Context, client *http.Client, primary, secondary, symbol string, out chan<- PriceTick) {
	const pollInterval = 5 * time.Second
	for {
		body, err := i.fetchPair(ctx, client, primary, secondary)
		if err == nil {
			tick, perr := i.parseSummary(body, symbol)
			if perr != nil {
				log.Printf("indep_reserve %s parse: %v", symbol, perr)
			} else if tick != nil {
				select {
				case <-ctx.Done():
					return
				case out <- *tick:
				default:
				}
			}
		} else if ctx.Err() == nil {
			log.Printf("indep_reserve %s fetch: %v", symbol, err)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(pollInterval):
		}
	}
}
