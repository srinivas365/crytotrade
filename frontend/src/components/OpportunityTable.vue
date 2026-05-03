<template>
  <div>
    <!-- Mobile: card layout -->
    <div class="block sm:hidden space-y-3">
      <div v-if="opportunities.length === 0" class="bg-white border border-gray-200 rounded-xl p-6 text-center text-gray-400 text-sm">
        No opportunities above your threshold right now.
      </div>
      <div
        v-for="opp in opportunities" :key="opp.symbol + opp.buyAt + opp.sellAt"
        class="bg-green-50 border border-green-200 border-l-4 border-l-green-400 rounded-xl p-4"
      >
        <div class="flex items-center justify-between mb-3">
          <span class="font-bold text-gray-900 text-base">{{ opp.symbol }}</span>
          <span class="text-sm font-bold text-green-600">+{{ opp.spreadPct.toFixed(4) }}%</span>
        </div>
        <div class="space-y-1.5 text-sm">
          <div class="flex justify-between">
            <span class="text-gray-500">Buy on <span class="capitalize font-medium text-gray-700">{{ opp.buyAt }}</span></span>
            <span class="text-gray-800 font-mono">${{ opp.buyPrice.toFixed(4) }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500">Sell on <span class="capitalize font-medium text-gray-700">{{ opp.sellAt }}</span></span>
            <span class="text-gray-800 font-mono">${{ opp.sellPrice.toFixed(4) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Desktop: table layout -->
    <div class="hidden sm:block bg-white rounded-xl border border-gray-200 overflow-hidden">
      <table class="w-full text-sm">
        <thead class="bg-gray-50 border-b border-gray-200">
          <tr>
            <th class="text-left px-4 py-3 font-semibold text-gray-600">Pair</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600">Buy At</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600">Buy Price</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600">Sell At</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600">Sell Price</th>
            <th class="text-right px-4 py-3 font-semibold text-gray-600">Spread %</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="opportunities.length === 0">
            <td colspan="6" class="text-center py-12 text-gray-400">No opportunities above your threshold right now.</td>
          </tr>
          <tr
            v-for="opp in opportunities" :key="opp.symbol + opp.buyAt + opp.sellAt"
            class="border-t border-gray-100 bg-green-50 border-l-4 border-l-green-400"
          >
            <td class="px-4 py-3 font-medium text-gray-900">{{ opp.symbol }}</td>
            <td class="px-4 py-3 capitalize text-gray-700">{{ opp.buyAt }}</td>
            <td class="px-4 py-3 text-gray-700">${{ opp.buyPrice.toFixed(4) }}</td>
            <td class="px-4 py-3 capitalize text-gray-700">{{ opp.sellAt }}</td>
            <td class="px-4 py-3 text-gray-700">${{ opp.sellPrice.toFixed(4) }}</td>
            <td class="px-4 py-3 text-right font-bold text-green-600">+{{ opp.spreadPct.toFixed(4) }}%</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { usePricesStore } from '@/stores/prices.js'
import { storeToRefs } from 'pinia'
const { opportunities } = storeToRefs(usePricesStore())
</script>
