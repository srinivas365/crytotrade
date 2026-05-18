package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var coindcxSymbols = map[string]string{
	// USDT pairs
	"BTCUSDT":   "BTC/USDT",
	"ETHUSDT":   "ETH/USDT",
	"SOLUSDT":   "SOL/USDT",
	"BNBUSDT":   "BNB/USDT",
	"XRPUSDT":   "XRP/USDT",
	"ADAUSDT":   "ADA/USDT",
	"DOGEUSDT":  "DOGE/USDT",
	"MATICUSDT": "MATIC/USDT",
	"LTCUSDT":   "LTC/USDT",
	"DOTUSDT":   "DOT/USDT",
	// INR pairs
	"BTCINR":   "BTC/INR",
	"ETHINR":   "ETH/INR",
	"SOLINR":   "SOL/INR",
	"XRPINR":   "XRP/INR",
	"ADAINR":   "ADA/INR",
	"DOGEINR":  "DOGE/INR",
	"MATICINR": "MATIC/INR",
	"LTCINR":   "LTC/INR",
	"DOTINR":   "DOT/INR",
}

type CoinDCX struct{}

// flexFloat accepts both JSON numbers (3.14) and JSON quoted strings ("3.14").
// CoinDCX's /exchange/ticker mostly returns quoted strings, but at least one
// market (BTCINR_insta) returns bare numbers, which would otherwise fail the
// whole array decode.
type flexFloat float64

func (f *flexFloat) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*f = flexFloat(v)
	return nil
}

type coindcxTicker struct {
	Market string    `json:"market"`
	Bid    flexFloat `json:"bid"`
	Ask    flexFloat `json:"ask"`
	Last   flexFloat `json:"last_price"`
}

func (c *CoinDCX) parseTickers(raw []byte) ([]PriceTick, error) {
	var arr []coindcxTicker
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, err
	}
	now := time.Now()
	var ticks []PriceTick
	for _, t := range arr {
		symbol, ok := coindcxSymbols[t.Market]
		if !ok {
			continue
		}
		bid, ask, last := float64(t.Bid), float64(t.Ask), float64(t.Last)
		if bid <= 0 {
			bid = last
		}
		if ask <= 0 {
			ask = last
		}
		if bid <= 0 || ask <= 0 {
			continue
		}
		ticks = append(ticks, PriceTick{
			Exchange:  "coindcx",
			Symbol:    symbol,
			Bid:       bid,
			Ask:       ask,
			Timestamp: now,
		})
	}
	return ticks, nil
}

func (c *CoinDCX) fetchOnce(ctx context.Context, client *http.Client) ([]PriceTick, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.coindcx.com/exchange/ticker", nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		snippet := strings.TrimSpace(string(body))
		if len(snippet) > 160 {
			snippet = snippet[:160] + "…"
		}
		return nil, fmt.Errorf("http %s: %s", resp.Status, snippet)
	}
	s := strings.TrimSpace(string(body))
	if len(s) > 0 && s[0] == '<' {
		prefix := s
		if len(prefix) > 80 {
			prefix = prefix[:80] + "…"
		}
		return nil, fmt.Errorf("response is HTML not JSON (proxy/WAF/captive portal?); prefix: %q", prefix)
	}
	return c.parseTickers(body)
}

func (c *CoinDCX) Connect(ctx context.Context, out chan<- PriceTick) {
	client := &http.Client{Timeout: 10 * time.Second}
	poll := time.NewTicker(3 * time.Second)
	defer poll.Stop()
	for {
		ticks, err := c.fetchOnce(ctx, client)
		if err != nil {
			log.Printf("coindcx fetch: %v — retrying in 5s", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}
		for _, tick := range ticks {
			select {
			case <-ctx.Done():
				return
			case out <- tick:
			default:
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-poll.C:
		}
	}
}
