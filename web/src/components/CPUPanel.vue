<template>
  <section class="panel">
    <div class="panel-hdr">
      <span class="panel-title">CPU</span>
      <span class="panel-sub" :class="cpu ? pctClass(cpu.usage_pct) : ''">
        {{ cpu ? cpu.usage_pct.toFixed(1) + '%' : '–' }}
      </span>
    </div>

    <div class="kv-row">
      <div class="kv">
        <span class="kv-k">Usage</span>
        <span class="kv-v" :class="cpu ? pctClass(cpu.usage_pct) : ''">
          {{ cpu ? cpu.usage_pct.toFixed(1) + '%' : '–' }}
        </span>
      </div>
      <div class="kv">
        <span class="kv-k">Temp</span>
        <span class="kv-v" :class="cpu ? tempClass(cpu.temp_c) : ''">
          {{ cpu ? cpu.temp_c.toFixed(0) + '°C' : '–' }}
        </span>
      </div>
      <div class="kv">
        <span class="kv-k">Freq</span>
        <span class="kv-v">
          {{ cpu ? (cpu.freq_mhz / 1000).toFixed(2) : '–' }}<span v-if="cpu" style="font-size:11px;color:var(--muted)"> GHz</span>
        </span>
      </div>
      <div class="kv">
        <span class="kv-k">Load avg</span>
        <span class="kv-v" style="font-size:13px;padding-top:3px">
          {{ cpu ? `${cpu.load1.toFixed(2)} ${cpu.load5.toFixed(2)} ${cpu.load15.toFixed(2)}` : '–' }}
        </span>
      </div>
    </div>

    <TimeChart ref="chart"
      :series="[{ color: '#58a6ff', fill: true }]"
      :max="100" :height="72" />
  </section>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { CPUStats } from '../types'
import { pctClass, tempClass } from '../utils'
import TimeChart from './TimeChart.vue'

const props = defineProps<{ cpu: CPUStats | null }>()
const chart = ref<InstanceType<typeof TimeChart>>()

watch(() => props.cpu?.usage_pct, v => { if (v != null) chart.value?.push(v) })
</script>

<style scoped>
.panel-hdr { display:flex; justify-content:space-between; align-items:baseline; }
.panel-title{ font-size:11px; font-weight:600; text-transform:uppercase; letter-spacing:.07em; color:var(--muted); }
.panel-sub  { font-size:11px; font-weight:700; }
.kv-row { display:flex; flex-wrap:wrap; gap:10px 20px; }
.kv     { display:flex; flex-direction:column; gap:1px; }
.kv-k   { font-size:10px; color:var(--muted); }
.kv-v   { font-size:16px; font-weight:700; line-height:1.1; }
</style>
