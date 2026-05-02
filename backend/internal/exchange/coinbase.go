package exchange

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var coinbaseSymbols = map[string]string{
	"BTC-USD":  "BTC/USDT",
	"ETH-USD":  "ETH/USDT",
	"SOL-USD":  "SOL/USDT",
	"XRP-USD":  "XRP/USDT",
	"ADA-USD":  "ADA/USDT",
	"DOGE-USD": "DOGE/USDT",
	"LTC-USD":  "LTC/USDT",
	"DOT-USD":  "DOT/USDT",
}

type Coinbase struct{}

type coinbaseMsg struct {
	Channel string `json:"channel"`
	Events  []struct {
		Type    string `json:"type"`
		Tickers []struct {
			ProductID string `json:"product_id"`
			BestBid   string `json:"best_bid"`
			BestAsk   string `json:"best_ask"`
		} `json:"tickers"`
	} `json:"events"`
}

func (c *Coinbase) parseMessage(raw []byte) ([]PriceTick, error) {
	var msg coinbaseMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, err
	}
	if msg.Channel != "ticker" {
		return nil, nil
	}
	var ticks []PriceTick
	for _, ev := range msg.Events {
		for _, t := range ev.Tickers {
			symbol, ok := coinbaseSymbols[t.ProductID]
			if !ok {
				continue
			}
			bid, _ := strconv.ParseFloat(t.BestBid, 64)
			ask, _ := strconv.ParseFloat(t.BestAsk, 64)
			if bid == 0 && ask == 0 {
				continue
			}
			ticks = append(ticks, PriceTick{
				Exchange:  "coinbase",
				Symbol:    symbol,
				Bid:       bid,
				Ask:       ask,
				Timestamp: time.Now(),
			})
		}
	}
	return ticks, nil
}

func (c *Coinbase) Connect(ctx context.Context, out chan<- PriceTick) {
	url := "wss://advanced-trade-ws.coinbase.com/ws"
	products := make([]string, 0, len(coinbaseSymbols))
	for k := range coinbaseSymbols {
		products = append(products, k)
	}
	subscribe, _ := json.Marshal(map[string]any{
		"type":        "subscribe",
		"product_ids": products,
		"channel":     "ticker",
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
		if err != nil {
			log.Printf("coinbase dial: %v — retrying in 5s", err)
			time.Sleep(5 * time.Second)
			continue
		}
		conn.WriteMessage(websocket.TextMessage, subscribe)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("coinbase read: %v", err)
				conn.Close()
				break
			}
			ticks, err := c.parseMessage(msg)
			if err != nil || len(ticks) == 0 {
				continue
			}
			for _, tick := range ticks {
				select {
				case out <- tick:
				default:
				}
			}
		}
	}
}
