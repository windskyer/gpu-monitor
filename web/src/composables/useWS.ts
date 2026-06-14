import { onMounted, onUnmounted, ref } from 'vue'
import type { Snapshot } from '../types'

export function useWS() {
  const snap = ref<Snapshot | null>(null)
  const connected = ref(false)
  let ws: WebSocket | null = null
  let timer: ReturnType<typeof setTimeout> | null = null
  let delay = 1000
  let stopped = false

  function scheduleReconnect() {
    if (stopped || timer !== null) return
    timer = setTimeout(() => { timer = null; connect() }, delay)
    delay = Math.min(delay * 2, 30_000)
  }

  function connect() {
    if (stopped) return
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const base = import.meta.env.BASE_URL.replace(/\/$/, '')
    const token = localStorage.getItem('gpu_monitor_token')
    const qs = token ? `?token=${encodeURIComponent(token)}` : ''
    const url = `${proto}://${location.host}${base}/ws${qs}`

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

  // start only if token exists; otherwise wait for setToken()
  onMounted(() => {
    const token = localStorage.getItem('gpu_monitor_token')
    if (token) connect()
  })
  onUnmounted(() => {
    stopped = true
    if (timer !== null) { clearTimeout(timer); timer = null }
    ws?.close()
  })

  function setToken(t: string) {
    localStorage.setItem('gpu_monitor_token', t)
    if (ws) {
      ws.close()
      ws = null
    }
    connect()
  }

  function clearToken() {
    localStorage.removeItem('gpu_monitor_token')
    if (ws) { ws.close(); ws = null }
    connected.value = false
  }

  return { snap, connected, setToken, clearToken }
}
