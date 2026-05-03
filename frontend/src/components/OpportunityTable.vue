<template>
  <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
    <div class="overflow-x-auto">
      <table class="w-full text-sm min-w-[580px]">
        <thead class="bg-gray-50 border-b border-gray-200">
          <tr>
            <th class="text-left px-3 sm:px-4 py-3 font-semibold text-gray-600">Pair</th>
            <th class="text-left px-3 sm:px-4 py-3 font-semibold text-gray-600">Buy At</th>
            <th class="text-left px-3 sm:px-4 py-3 font-semibold text-gray-600">Buy Price</th>
            <th class="text-left px-3 sm:px-4 py-3 font-semibold text-gray-600">Sell At</th>
            <th class="text-left px-3 sm:px-4 py-3 font-semibold text-gray-600">Sell Price</th>
            <th class="text-right px-3 sm:px-4 py-3 font-semibold text-gray-600">Spread %</th>
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
            <td class="px-3 sm:px-4 py-3 font-medium text-gray-900">{{ opp.symbol }}</td>
            <td class="px-3 sm:px-4 py-3 capitalize text-gray-700">{{ opp.buyAt }}</td>
            <td class="px-3 sm:px-4 py-3 text-gray-700">${{ opp.buyPrice.toFixed(4) }}</td>
            <td class="px-3 sm:px-4 py-3 capitalize text-gray-700">{{ opp.sellAt }}</td>
            <td class="px-3 sm:px-4 py-3 text-gray-700">${{ opp.sellPrice.toFixed(4) }}</td>
            <td class="px-3 sm:px-4 py-3 text-right font-bold text-green-600">+{{ opp.spreadPct.toFixed(4) }}%</td>
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
