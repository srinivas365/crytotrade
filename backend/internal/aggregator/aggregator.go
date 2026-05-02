package aggregator

import (
	"sync"

	"github.com/cryptotrade/app/internal/exchange"
)

type Aggregator struct {
	mu    sync.RWMutex
	ticks map[string]exchange.PriceTick // "exchange:symbol" → latest tick
}

func New() *Aggregator {
	return &Aggregator{ticks: make(map[string]exchange.PriceTick)}
}

func (a *Aggregator) UpdateTick(tick exchange.PriceTick) {
	key := tick.Exchange + ":" + tick.Symbol
	a.mu.Lock()
	a.ticks[key] = tick
	a.mu.Unlock()
}

func (a *Aggregator) GetTicks() map[string]exchange.PriceTick {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make(map[string]exchange.PriceTick, len(a.ticks))
	for k, v := range a.ticks {
		out[k] = v
	}
	return out
}

func (a *Aggregator) ComputeOpportunities() []SpreadOpportunity {
	ticks := a.GetTicks()
	bySymbol := make(map[string][]exchange.PriceTick)
	for _, t := range ticks {
		bySymbol[t.Symbol] = append(bySymbol[t.Symbol], t)
	}
	var opps []SpreadOpportunity
	for symbol, ts := range bySymbol {
		for i := range ts {
			for j := range ts {
				if i == j {
					continue
				}
				opp := CalculateSpread(symbol, ts[i], ts[j])
				if opp.SpreadPct > 0 {
					opps = append(opps, opp)
				}
			}
		}
	}
	return opps
}
