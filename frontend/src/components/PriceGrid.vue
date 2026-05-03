<template>
  <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
    <table class="w-full text-sm">
      <thead class="bg-gray-50 border-b border-gray-200">
        <tr>
          <th class="text-left px-4 py-3 font-semibold text-gray-600">Pair</th>
          <th class="text-left px-4 py-3 font-semibold text-gray-600">Binance</th>
          <th class="text-left px-4 py-3 font-semibold text-gray-600">Coinbase</th>
          <th class="text-left px-4 py-3 font-semibold text-gray-600">Kraken</th>
          <th
            class="text-right px-4 py-3 font-semibold text-gray-600 cursor-pointer select-none"
            @click="$emit('sort')"
          >Best Spread % ↕</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in rows" :key="row.symbol" class="border-t border-gray-100 hover:bg-gray-50">
          <td class="px-4 py-3 font-medium text-gray-900">{{ row.symbol }}</td>
          <td class="px-4 py-3 text-gray-700">{{ fmt(row.binance) }}</td>
          <td class="px-4 py-3 text-gray-700">{{ fmt(row.coinbase) }}</td>
          <td class="px-4 py-3 text-gray-700">{{ fmt(row.kraken) }}</td>
          <td class="px-4 py-3 text-right font-medium" :class="row.bestSpread > 0 ? 'text-green-600' : 'text-gray-400'">
            {{ row.bestSpread > 0 ? '+' + row.bestSpread.toFixed(4) + '%' : '—' }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { usePricesStore } from '@/stores/prices.js'
import { storeToRefs } from 'pinia'

const props = defineProps({ sortDesc: { type: Boolean, default: true } })
defineEmits(['sort'])

const { ticks } = storeToRefs(usePricesStore())

const EXCHANGES = ['binance', 'coinbase', 'kraken']

const rows = computed(() => {
  const bySymbol = new Map()
  for (const [, tick] of ticks.value) {
    if (!bySymbol.has(tick.symbol)) bySymbol.set(tick.symbol, {})
    bySymbol.get(tick.symbol)[tick.exchange] = tick
  }
  const result = []
  for (const [symbol, exMap] of bySymbol) {
    const ts = EXCHANGES.map(e => exMap[e]).filter(Boolean)
    let bestSpread = 0
    for (let i = 0; i < ts.length; i++) {
      for (let j = 0; j < ts.length; j++) {
        if (i === j) continue
        const sp = (ts[j].bid - ts[i].ask) / ts[i].ask * 100
        if (sp > bestSpread) bestSpread = sp
      }
    }
    result.push({ symbol, binance: exMap.binance, coinbase: exMap.coinbase, kraken: exMap.kraken, bestSpread })
  }
  return result.sort((a, b) => props.sortDesc ? b.bestSpread - a.bestSpread : a.bestSpread - b.bestSpread)
})

function fmt(tick) {
  if (!tick) return '—'
  return `$${tick.bid.toFixed(2)} / $${tick.ask.toFixed(2)}`
}
</script>
