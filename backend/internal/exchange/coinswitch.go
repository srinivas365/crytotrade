package exchange

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

// coinswitchSymbols maps the CoinSwitchX-quoted symbol (already canonical
// "BTC/INR" form) onto our normalised symbol. We restrict the universe here
// rather than emitting all ~400 pairs the endpoint returns.
var coinswitchSymbols = map[string]string{
	"BTC/INR":  "BTC/INR",
	"ETH/INR":  "ETH/INR",
	"SOL/INR":  "SOL/INR",
	"XRP/INR":  "XRP/INR",
	"ADA/INR":  "ADA/INR",
	"DOGE/INR": "DOGE/INR",
	"LTC/INR":  "LTC/INR",
	"DOT/INR":  "DOT/INR",
}

type CoinSwitch struct{}

type coinswitchTicker struct {
	BidPrice  string `json:"bidPrice"`
	AskPrice  string `json:"askPrice"`
	LastPrice string `json:"lastPrice"`
}

type coinswitchResp struct {
	Data map[string]coinswitchTicker `json:"data"`
}

func (c *CoinSwitch) parseTickers(raw []byte) ([]PriceTick, error) {
	var resp coinswitchResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	now := time.Now()
	var ticks []PriceTick
	for key, t := range resp.Data {
		symbol, ok := coinswitchSymbols[key]
		if !ok {
			continue
		}
		bid, _ := strconv.ParseFloat(t.BidPrice, 64)
		ask, _ := strconv.ParseFloat(t.AskPrice, 64)
		if bid <= 0 || ask <= 0 {
			last, _ := strconv.ParseFloat(t.LastPrice, 64)
			if last <= 0 {
				continue
			}
			if bid <= 0 {
				bid = last
			}
			if ask <= 0 {
				ask = last
			}
		}
		ticks = append(ticks, PriceTick{
			Exchange:  "coinswitch",
			Symbol:    symbol,
			Bid:       bid,
			Ask:       ask,
			Timestamp: now,
		})
	}
	return ticks, nil
}

func (c *CoinSwitch) fetchOnce(ctx context.Context, client *http.Client) ([]PriceTick, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://coinswitch.co/trade/api/v2/24hr/all-pairs/ticker?exchange=coinswitchx", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return c.parseTickers(body)
}

func (c *CoinSwitch) Connect(ctx context.Context, out chan<- PriceTick) {
	client := &http.Client{Timeout: 10 * time.Second}
	poll := time.NewTicker(3 * time.Second)
	defer poll.Stop()
	for {
		ticks, err := c.fetchOnce(ctx, client)
		if err != nil {
			log.Printf("coinswitch fetch: %v — retrying in 5s", err)
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
