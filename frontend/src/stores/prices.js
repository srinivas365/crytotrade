import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAuthStore } from './auth.js'

export const usePricesStore = defineStore('prices', () => {
  const ticks = ref(new Map())
  const alertQueue = ref([])

  const opportunities = computed(() => {
    const auth = useAuthStore()
    const threshold = auth.settings?.threshold_pct ?? 0.1

    const bySymbol = new Map()
    for (const [, tick] of ticks.value) {
      if (!bySymbol.has(tick.symbol)) bySymbol.set(tick.symbol, [])
      bySymbol.get(tick.symbol).push(tick)
    }

    const opps = []
    for (const [symbol, ts] of bySymbol) {
      for (let i = 0; i < ts.length; i++) {
        for (let j = 0; j < ts.length; j++) {
          if (i === j) continue
          const buyPrice = ts[i].ask
          const sellPrice = ts[j].bid
          if (!buyPrice || !sellPrice) continue
          const spreadPct = (sellPrice - buyPrice) / buyPrice * 100
          if (spreadPct >= threshold) {
            opps.push({ symbol, buyAt: ts[i].exchange, sellAt: ts[j].exchange, buyPrice, sellPrice, spreadPct })
          }
        }
      }
    }
    return opps.sort((a, b) => b.spreadPct - a.spreadPct)
  })

  function applySnapshot(snapshot) {
    const next = new Map()
    for (const [key, tick] of Object.entries(snapshot.ticks || {})) {
      next.set(key, tick)
    }
    ticks.value = next
  }

  function pushAlert(opp) { alertQueue.value.push(opp) }
  function shiftAlert() { return alertQueue.value.shift() }

  return { ticks, alertQueue, opportunities, applySnapshot, pushAlert, shiftAlert }
})
