package exchange

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var binanceSymbols = map[string]string{
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
}

type Binance struct{}

type binanceStreamMsg struct {
	Stream string `json:"stream"`
	Data   struct {
		Symbol      string `json:"s"`
		Bid         string `json:"b"`
		BidQuantity string `json:"B"`
		Ask         string `json:"a"`
		AskQuantity string `json:"A"`
	} `json:"data"`
}

func (b *Binance) parseMessage(raw []byte) (*PriceTick, error) {
	var msg binanceStreamMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, err
	}
	symbol, ok := binanceSymbols[msg.Data.Symbol]
	if !ok {
		return nil, nil
	}
	bid, _ := strconv.ParseFloat(msg.Data.Bid, 64)
	ask, _ := strconv.ParseFloat(msg.Data.Ask, 64)
	return &PriceTick{
		Exchange:  "binance",
		Symbol:    symbol,
		Bid:       bid,
		Ask:       ask,
		Timestamp: time.Now(),
	}, nil
}

func (b *Binance) Connect(ctx context.Context, out chan<- PriceTick) {
	keys := make([]string, 0, len(binanceSymbols))
	for k := range binanceSymbols {
		keys = append(keys, strings.ToLower(k)+"@bookTicker")
	}
	url := "wss://stream.binance.com:9443/stream?streams=" + strings.Join(keys, "/")

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
		if err != nil {
			log.Printf("binance dial: %v — retrying in 5s", err)
			time.Sleep(5 * time.Second)
			continue
		}
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("binance read: %v", err)
				conn.Close()
				break
			}
			tick, err := b.parseMessage(msg)
			if err != nil || tick == nil {
				continue
			}
			select {
			case out <- *tick:
			default:
			}
		}
	}
}
