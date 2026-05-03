<template>
  <NavBar v-if="auth.isAuthenticated" />
  <router-view />
  <AlertToast />
</template>

<script setup>
import { watch } from 'vue'
import { useAuthStore } from '@/stores/auth.js'
import { useWebSocket } from '@/composables/useWebSocket.js'
import NavBar from '@/components/NavBar.vue'
import AlertToast from '@/components/AlertToast.vue'

const auth = useAuthStore()
const { connect, disconnect } = useWebSocket()

watch(() => auth.isAuthenticated, (authed) => {
  if (authed) connect()
  else disconnect()
}, { immediate: true })
</script>
