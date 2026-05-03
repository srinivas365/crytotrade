import { useAuthStore } from '@/stores/auth.js'
import { usePricesStore } from '@/stores/prices.js'

let ws = null
let reconnectTimer = null
let backoff = 1000

export function useWebSocket() {
  function connect() {
    const auth = useAuthStore()
    if (!auth.token) return

    ws = new WebSocket(`/ws?token=${auth.token}`)

    ws.onopen = () => { backoff = 1000 }

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data)
      const prices = usePricesStore()
      if (msg.type === 'snapshot') {
        prices.applySnapshot(msg)
      } else if (msg.type === 'alert') {
        prices.pushAlert(msg.opportunity)
      }
    }

    ws.onclose = () => { scheduleReconnect() }
    ws.onerror = () => { ws?.close() }
  }

  function scheduleReconnect() {
    if (reconnectTimer) return
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      backoff = Math.min(backoff * 2, 30000)
      connect()
    }, backoff)
  }

  function disconnect() {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
    ws?.close()
    ws = null
  }

  return { connect, disconnect }
}
