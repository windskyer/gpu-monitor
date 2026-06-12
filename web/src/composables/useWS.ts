import { ref, onMounted, onUnmounted } from 'vue'
import type { Snapshot } from '../types'

export function useWS() {
  const snap      = ref<Snapshot | null>(null)
  const connected = ref(false)
  let ws: WebSocket | null = null
  let delay = 1000
  let stopped = false

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
    ws.onclose = ws.onerror = () => {
      connected.value = false
      ws = null
      if (!stopped) setTimeout(connect, delay)
      delay = Math.min(delay * 2, 30_000)
    }
  }

  onMounted(connect)
  onUnmounted(() => { stopped = true; ws?.close() })

  return { snap, connected }
}
