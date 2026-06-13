import { ref, onMounted, onUnmounted } from 'vue'
import type { Snapshot } from '../types'

export function useWS() {
  const snap      = ref<Snapshot | null>(null)
  const connected = ref(false)
  let ws:    WebSocket | null = null
  let timer: ReturnType<typeof setTimeout> | null = null
  let delay  = 1000
  let stopped = false

  function scheduleReconnect() {
    if (stopped || timer !== null) return
    timer = setTimeout(() => { timer = null; connect() }, delay)
    delay = Math.min(delay * 2, 30_000)
  }

  function connect() {
    if (stopped) return
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const tok   = new URLSearchParams(location.search).get('token') ?? ''
    const url   = `${proto}://${location.host}/ws${tok ? '?token=' + encodeURIComponent(tok) : ''}`

    ws = new WebSocket(url)

    ws.onopen = () => {
      connected.value = true
      delay = 1000
    }
    ws.onmessage = ev => {
      try { snap.value = JSON.parse(ev.data) as Snapshot }
      catch { /* ignore malformed frames */ }
    }
    ws.onerror = () => {
      // onerror always fires before onclose on failure; let onclose drive reconnect
    }
    ws.onclose = () => {
      connected.value = false
      ws = null
      scheduleReconnect()
    }
  }

  onMounted(connect)
  onUnmounted(() => {
    stopped = true
    if (timer !== null) { clearTimeout(timer); timer = null }
    ws?.close()
  })

  return { snap, connected }
}
