<template>
  <nav class="bg-white border-b border-gray-200 px-6 py-3 flex items-center justify-between">
    <span class="font-bold text-indigo-600 text-lg">CryptoTrade</span>
    <div class="flex gap-6 text-sm font-medium">
      <RouterLink to="/dashboard" class="text-gray-600 hover:text-indigo-600" active-class="text-indigo-600">Dashboard</RouterLink>
      <RouterLink to="/prices" class="text-gray-600 hover:text-indigo-600" active-class="text-indigo-600">All Prices</RouterLink>
      <RouterLink to="/history" class="text-gray-600 hover:text-indigo-600" active-class="text-indigo-600">Alert History</RouterLink>
      <RouterLink to="/settings" class="text-gray-600 hover:text-indigo-600" active-class="text-indigo-600">Settings</RouterLink>
    </div>
    <button @click="auth.logout()" class="text-sm text-gray-500 hover:text-red-500">Logout</button>
  </nav>
</template>

<script setup>
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth.js'
const auth = useAuthStore()
</script>
