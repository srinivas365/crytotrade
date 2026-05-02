package exchange

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var krakenSymbols = map[string]string{
	"XBT/USDT":  "BTC/USDT",
	"BTC/USDT":  "BTC/USDT",
	"ETH/USDT":  "ETH/USDT",
	"SOL/USDT":  "SOL/USDT",
	"XRP/USDT":  "XRP/USDT",
	"ADA/USDT":  "ADA/USDT",
	"DOGE/USDT": "DOGE/USDT",
	"LTC/USDT":  "LTC/USDT",
	"DOT/USDT":  "DOT/USDT",
}

var krakenPairs = func() []string {
	seen := map[string]bool{}
	var pairs []string
	for _, v := range krakenSymbols {
		if !seen[v] {
			seen[v] = true
			pairs = append(pairs, v)
		}
	}
	return pairs
}()

type Kraken struct{}

type krakenMsg struct {
	Channel string `json:"channel"`
	Type    string `json:"type"`
	Data    []struct {
		Symbol string  `json:"symbol"`
		Bid    float64 `json:"bid"`
		Ask    float64 `json:"ask"`
	} `json:"data"`
}

func (k *Kraken) parseMessage(raw []byte) ([]PriceTick, error) {
	var msg krakenMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, err
	}
	if msg.Channel != "ticker" {
		return nil, nil
	}
	var ticks []PriceTick
	for _, d := range msg.Data {
		symbol, ok := krakenSymbols[d.Symbol]
		if !ok {
			continue
		}
		ticks = append(ticks, PriceTick{
			Exchange:  "kraken",
			Symbol:    symbol,
			Bid:       d.Bid,
			Ask:       d.Ask,
			Timestamp: time.Now(),
		})
	}
	return ticks, nil
}

func (k *Kraken) Connect(ctx context.Context, out chan<- PriceTick) {
	url := "wss://ws.kraken.com/v2"
	subscribe, _ := json.Marshal(map[string]any{
		"method": "subscribe",
		"params": map[string]any{
			"channel": "ticker",
			"symbol":  krakenPairs,
		},
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
		if err != nil {
			log.Printf("kraken dial: %v — retrying in 5s", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, subscribe); err != nil {
			log.Printf("kraken subscribe: %v", err)
			conn.Close()
			continue
		}
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("kraken read: %v", err)
				conn.Close()
				break
			}
			ticks, err := k.parseMessage(msg)
			if err != nil || len(ticks) == 0 {
				continue
			}
			for _, tick := range ticks {
				select {
				case <-ctx.Done():
					conn.Close()
					return
				case out <- tick:
				default:
				}
			}
		}
	}
}
