<template>
  <div class="min-h-screen bg-gray-50 p-6">
    <div class="flex items-center justify-between mb-5">
      <h2 class="text-xl font-bold text-gray-900">All Prices</h2>
      <span class="text-xs text-gray-400">Updates every 500ms</span>
    </div>
    <PriceGrid :sort-desc="sortDesc" @sort="sortDesc = !sortDesc" />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import PriceGrid from '@/components/PriceGrid.vue'
const sortDesc = ref(true)
</script>
