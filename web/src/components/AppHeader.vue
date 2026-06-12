<template>
  <header class="hdr">
    <span class="logo">&#9889; GPU Monitor</span>
    <span class="load" v-if="cpu">
      load {{ cpu.load1.toFixed(2) }}&nbsp;{{ cpu.load5.toFixed(2) }}&nbsp;{{ cpu.load15.toFixed(2) }}
    </span>
    <div class="right">
      <span class="ts">{{ ts }}</span>
      <span class="dot" :class="connected ? 'online' : 'offline'" />
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { CPUStats } from '../types'

const props = defineProps<{
  connected: boolean
  ts:        string
  cpu?:      CPUStats
}>()

// suppress unused warning — ts is used in template
void props.ts
</script>

<style scoped>
.hdr   { display:flex; align-items:center; justify-content:space-between; gap:12px;
         padding:7px 14px; background:var(--surface); border-bottom:1px solid var(--border);
         position:sticky; top:0; z-index:100; }
.logo  { font-size:13px; font-weight:700; color:var(--blue); white-space:nowrap; }
.load  { flex:1; text-align:center; font-size:11px; color:var(--muted); white-space:nowrap;
         overflow:hidden; text-overflow:ellipsis; }
.right { display:flex; align-items:center; gap:8px; white-space:nowrap; }
.ts    { font-size:11px; color:var(--muted); }
.dot   { width:8px; height:8px; border-radius:50%; flex-shrink:0; transition:background .3s; }
.dot.online  { background:var(--green); animation:blink 2.5s ease-in-out infinite; }
.dot.offline { background:var(--red); }
@keyframes blink { 0%,100%{opacity:1} 50%{opacity:.4} }
</style>
