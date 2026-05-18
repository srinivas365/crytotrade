<template>
  <div>
    <!-- Mobile: card layout (hidden on sm+) -->
    <div class="block sm:hidden">
      <div class="flex items-center justify-between mb-3">
        <span class="text-xs text-gray-500">{{ rows.length }} pairs</span>
        <button
          @click="$emit('sort')"
          class="text-xs font-medium text-indigo-600 border border-indigo-200 rounded-lg px-3 py-1"
        >Sort by Spread ↕</button>
      </div>
      <div class="space-y-3">
        <div
          v-for="row in rows" :key="row.symbol"
          class="bg-white border border-gray-200 rounded-xl p-4"
        >
          <div class="flex items-center justify-between mb-3">
            <span class="font-bold text-gray-900 text-base">{{ row.symbol }}</span>
            <span
              class="text-sm font-bold"
              :class="row.bestSpread > 0 ? 'text-green-600' : 'text-gray-400'"
            >{{ row.bestSpread > 0 ? '+' + row.bestSpread.toFixed(4) + '%' : '—' }}</span>
          </div>
          <div class="space-y-1.5 text-sm">
            <div v-for="ex in EXCHANGES" :key="ex.id" class="flex justify-between">
              <span class="text-gray-500 w-20">{{ ex.label }}</span>
              <span class="text-gray-800 font-mono">{{ fmt(row.byExchange[ex.id]) }}</span>
            </div>
          </div>
        </div>
        <p v-if="rows.length === 0" class="text-center py-8 text-gray-400 text-sm">No price data yet.</p>
      </div>
    </div>

    <!-- Desktop: table layout (hidden on mobile) -->
    <div class="hidden sm:block bg-white rounded-xl border border-gray-200 overflow-hidden">
      <table class="w-full text-sm">
        <thead class="bg-gray-50 border-b border-gray-200">
          <tr>
            <th class="text-left px-4 py-3 font-semibold text-gray-600">Pair</th>
            <th
              v-for="ex in EXCHANGES" :key="ex.id"
              class="text-left px-4 py-3 font-semibold text-gray-600"
            >{{ ex.label }}</th>
            <th
              class="text-right px-4 py-3 font-semibold text-gray-600 cursor-pointer select-none"
              @click="$emit('sort')"
            >Best Spread % ↕</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in rows" :key="row.symbol" class="border-t border-gray-100 hover:bg-gray-50">
            <td class="px-4 py-3 font-medium text-gray-900">{{ row.symbol }}</td>
            <td
              v-for="ex in EXCHANGES" :key="ex.id"
              class="px-4 py-3 text-gray-700"
            >{{ fmt(row.byExchange[ex.id]) }}</td>
            <td class="px-4 py-3 text-right font-medium" :class="row.bestSpread > 0 ? 'text-green-600' : 'text-gray-400'">
              {{ row.bestSpread > 0 ? '+' + row.bestSpread.toFixed(4) + '%' : '—' }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { usePricesStore } from '@/stores/prices.js'
import { storeToRefs } from 'pinia'

const props = defineProps({ sortDesc: { type: Boolean, default: true } })
defineEmits(['sort'])

const { ticks } = storeToRefs(usePricesStore())

const EXCHANGES = [
  { id: 'binance',  label: 'Binance' },
  { id: 'coinbase', label: 'Coinbase' },
  { id: 'kraken',   label: 'Kraken' },
  { id: 'coindcx',       label: 'CoinDCX' },
  { id: 'coinswitch',    label: 'CoinSwitch' },
  { id: 'indep_reserve', label: 'Indep Reserve' },
]

const rows = computed(() => {
  const bySymbol = new Map()
  for (const [, tick] of ticks.value) {
    if (!bySymbol.has(tick.symbol)) bySymbol.set(tick.symbol, {})
    bySymbol.get(tick.symbol)[tick.exchange] = tick
  }
  const result = []
  for (const [symbol, exMap] of bySymbol) {
    const ts = EXCHANGES.map(e => exMap[e.id]).filter(Boolean)
    let bestSpread = 0
    for (let i = 0; i < ts.length; i++) {
      for (let j = 0; j < ts.length; j++) {
        if (i === j) continue
        const sp = (ts[j].bid - ts[i].ask) / ts[i].ask * 100
        if (sp > bestSpread) bestSpread = sp
      }
    }
    result.push({ symbol, byExchange: exMap, bestSpread })
  }
  return result.sort((a, b) => props.sortDesc ? b.bestSpread - a.bestSpread : a.bestSpread - b.bestSpread)
})

function fmt(tick) {
  if (!tick) return '—'
  return `$${tick.bid.toFixed(2)} / $${tick.ask.toFixed(2)}`
}
</script>
