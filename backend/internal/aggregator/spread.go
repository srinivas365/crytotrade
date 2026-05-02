package aggregator

import (
	"time"

	"github.com/cryptotrade/app/internal/exchange"
)

type SpreadOpportunity struct {
	Symbol     string    `json:"symbol"`
	BuyAt      string    `json:"buy_at"`
	SellAt     string    `json:"sell_at"`
	BuyPrice   float64   `json:"buy_price"`
	SellPrice  float64   `json:"sell_price"`
	SpreadPct  float64   `json:"spread_pct"`
	DetectedAt time.Time `json:"detected_at"`
}

// CalculateSpread returns the spread when buying at buyTick.Ask and selling at sellTick.Bid.
func CalculateSpread(symbol string, buyTick, sellTick exchange.PriceTick) SpreadOpportunity {
	buyPrice := buyTick.Ask
	sellPrice := sellTick.Bid
	spreadPct := 0.0
	if buyPrice > 0 {
		spreadPct = (sellPrice - buyPrice) / buyPrice * 100
	}
	return SpreadOpportunity{
		Symbol:     symbol,
		BuyAt:      buyTick.Exchange,
		SellAt:     sellTick.Exchange,
		BuyPrice:   buyPrice,
		SellPrice:  sellPrice,
		SpreadPct:  spreadPct,
		DetectedAt: time.Now(),
	}
}
