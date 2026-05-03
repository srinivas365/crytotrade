<template>
  <div class="min-h-screen bg-gray-50 p-6">
    <h2 class="text-xl font-bold text-gray-900 mb-5">Alert History</h2>
    <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
      <table class="w-full text-sm">
        <thead class="bg-gray-50 border-b border-gray-200">
          <tr>
            <th class="text-left px-4 py-3 font-semibold text-gray-600">Time</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600">Pair</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600">Buy</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600">Sell</th>
            <th class="text-right px-4 py-3 font-semibold text-gray-600">Spread %</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="!records.length">
            <td colspan="5" class="text-center py-12 text-gray-400">No alerts yet.</td>
          </tr>
          <tr v-for="r in records" :key="r.id" class="border-t border-gray-100 hover:bg-gray-50">
            <td class="px-4 py-3 text-gray-500">{{ fmtDate(r.fired_at) }}</td>
            <td class="px-4 py-3 font-medium text-gray-900">{{ r.symbol }}</td>
            <td class="px-4 py-3 capitalize text-gray-700">{{ r.buy_exchange }} ${{ r.buy_price.toFixed(4) }}</td>
            <td class="px-4 py-3 capitalize text-gray-700">{{ r.sell_exchange }} ${{ r.sell_price.toFixed(4) }}</td>
            <td class="px-4 py-3 text-right font-bold text-green-600">+{{ r.spread_pct.toFixed(4) }}%</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="totalPages > 1" class="flex items-center justify-between mt-4">
      <span class="text-sm text-gray-500">Page {{ page }} of {{ totalPages }} ({{ total }} total)</span>
      <div class="flex gap-2">
        <button
          @click="fetchPage(page - 1)" :disabled="page === 1"
          class="px-3 py-1.5 text-sm rounded-lg border border-gray-300 text-gray-600 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
        >← Prev</button>
        <button
          @click="fetchPage(page + 1)" :disabled="page === totalPages"
          class="px-3 py-1.5 text-sm rounded-lg border border-gray-300 text-gray-600 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
        >Next →</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth.js'

const auth = useAuthStore()
const records = ref([])
const total = ref(0)
const page = ref(1)
const limit = 20

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / limit)))

async function fetchPage(p) {
  page.value = p
  const res = await fetch(`/api/history?page=${p}&limit=${limit}`, {
    headers: { Authorization: `Bearer ${auth.token}` },
  })
  if (res.ok) {
    const data = await res.json()
    records.value = data.records
    total.value = data.total
  }
}

onMounted(() => fetchPage(1))

function fmtDate(iso) {
  return new Date(iso).toLocaleString()
}
</script>
