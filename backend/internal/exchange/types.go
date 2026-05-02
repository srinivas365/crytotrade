package exchange

import "time"

type PriceTick struct {
	Exchange  string    `json:"exchange"`
	Symbol    string    `json:"symbol"` // normalised: "BTC/USDT"
	Bid       float64   `json:"bid"`
	Ask       float64   `json:"ask"`
	Timestamp time.Time `json:"timestamp"`
}

// TickHandler receives normalised ticks from a connector.
type TickHandler func(PriceTick)
