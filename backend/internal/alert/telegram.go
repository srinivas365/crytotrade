package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cryptotrade/app/internal/aggregator"
)

func SendTelegram(botToken, chatID string, opp aggregator.SpreadOpportunity) error {
	text := fmt.Sprintf(
		"Arbitrage Opportunity!\n\nPair: %s\nBuy on %s at $%.4f\nSell on %s at $%.4f\nSpread: +%.4f%%",
		opp.Symbol, opp.BuyAt, opp.BuyPrice, opp.SellAt, opp.SellPrice, opp.SpreadPct,
	)
	payload, _ := json.Marshal(map[string]string{"chat_id": chatID, "text": text})
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
