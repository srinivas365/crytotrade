<template>
  <Teleport to="body">
    <div class="fixed bottom-4 right-4 left-4 sm:left-auto sm:right-5 sm:bottom-5 space-y-2 z-50">
      <TransitionGroup name="toast">
        <div
          v-for="toast in toasts" :key="toast.id"
          class="bg-white border border-green-300 shadow-lg rounded-xl p-4 w-full sm:w-80"
        >
          <p class="font-semibold text-green-700 text-sm">Opportunity: {{ toast.symbol }}</p>
          <p class="text-xs text-gray-600 mt-1">
            Buy {{ toast.buy_at }} ${{ toast.buy_price.toFixed(4) }} →
            Sell {{ toast.sell_at }} ${{ toast.sell_price.toFixed(4) }}
          </p>
          <p class="text-xs font-bold text-green-600 mt-1">+{{ toast.spread_pct.toFixed(4) }}%</p>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, watch } from 'vue'
import { usePricesStore } from '@/stores/prices.js'
import { useAuthStore } from '@/stores/auth.js'

const prices = usePricesStore()
const toasts = ref([])
let id = 0

watch(() => prices.alertQueue.length, () => {
  const opp = prices.shiftAlert()
  if (!opp) return
  const toast = { ...opp, id: ++id }
  toasts.value.push(toast)
  const auth = useAuthStore()
  if (auth.settings?.alert_sound) {
    const ctx = new AudioContext()
    const osc = ctx.createOscillator()
    osc.connect(ctx.destination)
    osc.frequency.value = 440
    osc.start()
    osc.stop(ctx.currentTime + 0.15)
  }
  setTimeout(() => {
    toasts.value = toasts.value.filter(t => t.id !== toast.id)
  }, 6000)
})
</script>

<style scoped>
.toast-enter-active, .toast-leave-active { transition: all 0.3s ease; }
.toast-enter-from, .toast-leave-to { opacity: 0; transform: translateX(40px); }
</style>
