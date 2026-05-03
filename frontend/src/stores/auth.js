import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import router from '@/router/index.js'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userID = ref(localStorage.getItem('user_id') || '')
  const settings = ref(null)

  const isAuthenticated = computed(() => !!token.value)

  async function _setSession(data) {
    token.value = data.token
    userID.value = data.user_id
    localStorage.setItem('token', data.token)
    localStorage.setItem('user_id', data.user_id)
    await fetchSettings()
  }

  async function login(email, password) {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
    if (!res.ok) throw new Error('Invalid credentials')
    await _setSession(await res.json())
    router.push('/dashboard')
  }

  async function register(email, password) {
    const res = await fetch('/api/auth/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
    if (!res.ok) throw new Error((await res.text()) || 'Registration failed')
    await _setSession(await res.json())
    router.push('/dashboard')
  }

  async function fetchSettings() {
    const res = await fetch('/api/settings', {
      headers: { Authorization: `Bearer ${token.value}` },
    })
    if (res.ok) settings.value = await res.json()
  }

  async function updateSettings(s) {
    const res = await fetch('/api/settings', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token.value}` },
      body: JSON.stringify(s),
    })
    if (!res.ok) throw new Error('Failed to save')
    settings.value = { ...settings.value, ...s }
  }

  function logout() {
    token.value = ''
    userID.value = ''
    settings.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user_id')
    router.push('/login')
  }

  return { token, userID, settings, isAuthenticated, login, register, logout, fetchSettings, updateSettings }
})
