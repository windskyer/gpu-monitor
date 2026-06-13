<template>
  <AppHeader
    :connected="connected"
    :ts="ts"
    :cpu="snap?.cpu" />

  <div v-if="showTokenPrompt" class="token-overlay">
    <div class="token-box panel">
      <h3>请输入访问 Token</h3>
      <input v-model="tokenInput" placeholder="输入 token" />
      <div class="actions">
        <button @click="submitToken">进入</button>
      </div>
      <p class="hint">Token 会保存在本地存储，用于后续自动登录。</p>
    </div>
  </div>

  <main class="app">
    <!-- GPU panels -->
    <template v-if="snap?.gpus?.length">
      <GPUPanel v-for="g in snap.gpus" :key="g.index" :g="g" />
    </template>
    <div v-else class="panel no-gpu">
      {{ connected ? 'No GPUs detected' : 'Connecting…' }}
    </div>

    <!-- System row -->
    <div class="lower-grid">
      <CPUPanel :cpu="snap?.cpu ?? null" />
      <MemPanel :mem="snap?.mem ?? null" />
      <NetPanel :nets="snap?.networks ?? []" />
      <DiskPanel :disks="snap?.disks ?? []" />
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import AppHeader from './components/AppHeader.vue'
import CPUPanel from './components/CPUPanel.vue'
import DiskPanel from './components/DiskPanel.vue'
import GPUPanel from './components/GPUPanel.vue'
import MemPanel from './components/MemPanel.vue'
import NetPanel from './components/NetPanel.vue'
import { useWS } from './composables/useWS'

const { snap, connected, setToken, clearToken } = useWS()

const tokenInput = ref(localStorage.getItem('gpu_monitor_token') || '')
const showTokenPrompt = computed(() => !connected.value)

function submitToken() {
  if (!tokenInput.value) return
  setToken(tokenInput.value)
}

const ts = computed(() => snap.value
  ? new Date(snap.value.ts).toLocaleString('zh-CN', { hour12: false })
  : '–')
</script>

<style>
/* ── Global tokens ─────────────────────────────────────────────────────────── */
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
html, body { height: 100%; }

:root {
  --bg:      #0d1117;
  --surface: #161b22;
  --raised:  #1c2128;
  --border:  #21262d;
  --border2: #30363d;
  --text:    #e6edf3;
  --muted:   #8b949e;
  --dim:     #484f58;
  --green:   #3fb950;
  --cyan:    #39d353;
  --blue:    #58a6ff;
  --purple:  #bc8cff;
  --yellow:  #d29922;
  --orange:  #f0883e;
  --red:     #f85149;
  --r: 6px;
}

body {
  background: var(--bg);
  color: var(--text);
  font-family: 'SF Mono', ui-monospace, 'Cascadia Code', 'Consolas', monospace;
  font-size: 12px;
  line-height: 1.4;
  overflow-x: hidden;
}

/* ── Utility color classes ─────────────────────────────────────────────────── */
.c-green  { color: var(--green) !important; }
.c-cyan   { color: var(--cyan) !important; }
.c-blue   { color: var(--blue) !important; }
.c-yellow { color: var(--yellow) !important; }
.c-orange { color: var(--orange) !important; }
.c-red    { color: var(--red) !important; }
.c-purple { color: var(--purple) !important; }

/* ── Layout ────────────────────────────────────────────────────────────────── */
.app { display: flex; flex-direction: column; gap: 8px; padding: 8px; }

.no-gpu {
  color: var(--dim);
  font-size: 12px;
  text-align: center;
  padding: 20px;
}

.lower-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
  /* Let rows size to content to avoid large vertical whitespace */
  grid-auto-rows: auto;
  align-items: start;
}

@media (max-width: 600px) {
  .lower-grid { grid-template-columns: 1fr; }
}

/* Ensure direct grid children can shrink/grow properly */
.lower-grid > * { min-height: 0; }

/* ── Panel base ────────────────────────────────────────────────────────────── */
.panel {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r);
  padding: 8px 10px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  overflow: hidden;
}

.token-overlay {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(5,8,10,0.6);
  z-index: 60;
}
.token-box { width: 420px; max-width: calc(100% - 24px); padding: 16px; }
.token-box h3 { margin-bottom: 8px; color: var(--text); }
.token-box input { width: 100%; padding: 8px; border-radius: 6px; border: 1px solid var(--border2); background: var(--raised); color: var(--text); }
.token-box .actions { margin-top: 8px; display:flex; justify-content:flex-end }
.token-box button { padding: 8px 12px; border-radius:6px; border:0; background:var(--green); color:#06120a; font-weight:700 }
.token-box .hint { margin-top:8px; color:var(--dim); font-size:12px }

/* ── Scrollbar ─────────────────────────────────────────────────────────────── */
::-webkit-scrollbar { width: 6px; height: 6px; }
::-webkit-scrollbar-track  { background: var(--bg); }
::-webkit-scrollbar-thumb  { background: var(--border2); border-radius: 3px; }
::-webkit-scrollbar-thumb:hover { background: var(--muted); }
</style>
