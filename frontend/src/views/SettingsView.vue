<template>
  <div class="min-h-screen bg-gray-50 p-4 sm:p-6">
    <h2 class="text-xl font-bold text-gray-900 mb-5">Settings</h2>
    <div class="bg-white rounded-xl border border-gray-200 p-4 sm:p-6 w-full sm:max-w-lg">
      <form @submit.prevent="save" class="space-y-5">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Alert Threshold (%)</label>
          <input v-model.number="form.threshold_pct" type="number" min="0" max="100" step="0.01"
            class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
          <p class="text-xs text-gray-400 mt-1">Minimum spread % to trigger an alert</p>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Telegram Bot Token</label>
          <input v-model="form.telegram_bot_token" type="text" placeholder="123456:ABC..."
            class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Telegram Chat ID</label>
          <input v-model="form.telegram_chat_id" type="text" placeholder="-1001234567890"
            class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
        </div>
        <div class="flex items-center justify-between">
          <label class="text-sm font-medium text-gray-700">In-app alerts</label>
          <button type="button" @click="form.in_app_alerts = !form.in_app_alerts"
            :class="form.in_app_alerts ? 'bg-indigo-600' : 'bg-gray-300'"
            class="relative inline-flex w-11 h-6 rounded-full transition-colors duration-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500">
            <span :class="form.in_app_alerts ? 'translate-x-5' : 'translate-x-0.5'"
              class="absolute top-0.5 left-0 w-5 h-5 bg-white rounded-full shadow transition-transform duration-200"></span>
          </button>
        </div>
        <div class="flex items-center justify-between">
          <label class="text-sm font-medium text-gray-700">Alert sound</label>
          <button type="button" @click="form.alert_sound = !form.alert_sound"
            :class="form.alert_sound ? 'bg-indigo-600' : 'bg-gray-300'"
            class="relative inline-flex w-11 h-6 rounded-full transition-colors duration-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500">
            <span :class="form.alert_sound ? 'translate-x-5' : 'translate-x-0.5'"
              class="absolute top-0.5 left-0 w-5 h-5 bg-white rounded-full shadow transition-transform duration-200"></span>
          </button>
        </div>
        <p v-if="saved" class="text-green-600 text-sm">Settings saved.</p>
        <p v-if="err" class="text-red-500 text-sm">{{ err }}</p>
        <button type="submit"
          class="w-full bg-indigo-600 text-white py-2 rounded-lg text-sm font-medium hover:bg-indigo-700">
          Save Settings
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth.js'

const auth = useAuthStore()
const saved = ref(false)
const err = ref('')
const form = ref({
  threshold_pct: 0.1,
  telegram_bot_token: '',
  telegram_chat_id: '',
  in_app_alerts: true,
  alert_sound: true,
})

onMounted(() => {
  if (auth.settings) Object.assign(form.value, auth.settings)
})

async function save() {
  saved.value = false
  err.value = ''
  try {
    await auth.updateSettings(form.value)
    saved.value = true
    setTimeout(() => { saved.value = false }, 3000)
  } catch (e) {
    err.value = e.message
  }
}
</script>
