<template>
  <div class="min-h-screen bg-gray-50 p-6">
    <h2 class="text-xl font-bold text-gray-900 mb-5">Dashboard</h2>

    <div class="grid grid-cols-3 gap-4 mb-6">
      <StatCard
        label="Active Opportunities"
        :value="opportunities.length"
        value-class="text-green-600"
        :sub="`threshold: ${threshold}%`"
      />
      <StatCard
        label="Avg Spread"
        :value="avgSpread ? avgSpread + '%' : '—'"
        value-class="text-gray-900"
      />
      <StatCard
        label="Tracked Pairs"
        :value="trackedPairs"
        value-class="text-gray-900"
      />
    </div>

    <OpportunityTable />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { usePricesStore } from '@/stores/prices.js'
import { useAuthStore } from '@/stores/auth.js'
import { storeToRefs } from 'pinia'
import StatCard from '@/components/StatCard.vue'
import OpportunityTable from '@/components/OpportunityTable.vue'

const { opportunities, ticks } = storeToRefs(usePricesStore())
const auth = useAuthStore()
const threshold = computed(() => auth.settings?.threshold_pct ?? 0.1)
const avgSpread = computed(() => {
  if (!opportunities.value.length) return null
  const avg = opportunities.value.reduce((s, o) => s + o.spreadPct, 0) / opportunities.value.length
  return avg.toFixed(4)
})
const trackedPairs = computed(() => new Set([...ticks.value.values()].map(t => t.symbol)).size)
</script>
