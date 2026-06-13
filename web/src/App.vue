<template>
  <AppHeader
    :connected="connected"
    :ts="ts"
    :cpu="snap?.cpu" />

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
import { computed } from 'vue'
import AppHeader from './components/AppHeader.vue'
import CPUPanel from './components/CPUPanel.vue'
import DiskPanel from './components/DiskPanel.vue'
import GPUPanel from './components/GPUPanel.vue'
import MemPanel from './components/MemPanel.vue'
import NetPanel from './components/NetPanel.vue'
import { useWS } from './composables/useWS'

const { snap, connected } = useWS()

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
  gap: 8px;
  /* Make rows equal height so panels align vertically */
  grid-auto-rows: 1fr;
  align-items: stretch;
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
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow: hidden;
}

/* ── Scrollbar ─────────────────────────────────────────────────────────────── */
::-webkit-scrollbar { width: 6px; height: 6px; }
::-webkit-scrollbar-track  { background: var(--bg); }
::-webkit-scrollbar-thumb  { background: var(--border2); border-radius: 3px; }
::-webkit-scrollbar-thumb:hover { background: var(--muted); }
</style>
