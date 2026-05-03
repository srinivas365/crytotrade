<template>
  <div class="min-h-screen bg-gray-50 flex items-center justify-center">
    <div class="bg-white rounded-xl shadow p-8 w-full max-w-sm">
      <h1 class="text-2xl font-bold text-gray-900 mb-6 text-center">CryptoTrade</h1>

      <div class="flex rounded-lg border border-gray-200 mb-6">
        <button
          v-for="tab in ['Login', 'Register']" :key="tab"
          @click="mode = tab"
          class="flex-1 py-2 text-sm font-medium rounded-lg transition"
          :class="mode === tab ? 'bg-indigo-600 text-white' : 'text-gray-600 hover:bg-gray-50'"
        >{{ tab }}</button>
      </div>

      <form @submit.prevent="submit" class="space-y-4">
        <input
          v-model="email" type="email" placeholder="Email" required
          class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
        <input
          v-model="password" type="password" placeholder="Password (min 8 chars)" required
          class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
        <p v-if="error" class="text-red-500 text-sm">{{ error }}</p>
        <button
          type="submit" :disabled="loading"
          class="w-full bg-indigo-600 text-white py-2 rounded-lg text-sm font-medium hover:bg-indigo-700 disabled:opacity-50"
        >{{ loading ? 'Please wait...' : mode }}</button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth.js'

const auth = useAuthStore()
const mode = ref('Login')
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    if (mode.value === 'Login') await auth.login(email.value, password.value)
    else await auth.register(email.value, password.value)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>
