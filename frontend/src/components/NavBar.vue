<template>
  <nav class="bg-white border-b border-gray-200">
    <div class="px-4 sm:px-6 py-3 flex items-center justify-between">
      <span class="font-bold text-indigo-600 text-lg">CryptoTrade</span>

      <div class="hidden md:flex gap-6 text-sm font-medium">
        <RouterLink to="/dashboard" class="text-gray-600 hover:text-indigo-600" active-class="text-indigo-600">Dashboard</RouterLink>
        <RouterLink to="/prices" class="text-gray-600 hover:text-indigo-600" active-class="text-indigo-600">All Prices</RouterLink>
        <RouterLink to="/history" class="text-gray-600 hover:text-indigo-600" active-class="text-indigo-600">Alert History</RouterLink>
        <RouterLink to="/settings" class="text-gray-600 hover:text-indigo-600" active-class="text-indigo-600">Settings</RouterLink>
      </div>

      <div class="flex items-center gap-3">
        <button @click="auth.logout()" class="hidden md:block text-sm text-gray-500 hover:text-red-500">Logout</button>
        <button @click="open = !open" class="md:hidden p-1 text-gray-500 hover:text-indigo-600">
          <svg v-if="!open" class="w-6 h-6" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16"/>
          </svg>
          <svg v-else class="w-6 h-6" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>
    </div>

    <div v-if="open" class="md:hidden border-t border-gray-100 px-4 py-3 flex flex-col gap-4 text-sm font-medium">
      <RouterLink @click="open = false" to="/dashboard" class="text-gray-600 hover:text-indigo-600" active-class="text-indigo-600">Dashboard</RouterLink>
      <RouterLink @click="open = false" to="/prices" class="text-gray-600 hover:text-indigo-600" active-class="text-indigo-600">All Prices</RouterLink>
      <RouterLink @click="open = false" to="/history" class="text-gray-600 hover:text-indigo-600" active-class="text-indigo-600">Alert History</RouterLink>
      <RouterLink @click="open = false" to="/settings" class="text-gray-600 hover:text-indigo-600" active-class="text-indigo-600">Settings</RouterLink>
      <button @click="auth.logout()" class="text-left text-red-500 text-sm">Logout</button>
    </div>
  </nav>
</template>

<script setup>
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth.js'
const auth = useAuthStore()
const open = ref(false)
</script>
