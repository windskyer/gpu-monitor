<template>
  <section class="panel">
    <div class="panel-hdr">
      <span class="panel-title">Memory</span>
      <span class="panel-sub" :class="mem ? pctClass(mem.used_pct) : ''">
        {{ mem ? mem.used_pct.toFixed(1) + '%' : '–' }}
      </span>
    </div>

    <template v-if="mem">
      <ProgressBar label="RAM" :value="mem.used_pct"
        :right="`${fmtBytes(mem.used)} / ${fmtBytes(mem.total)}`" />
      <ProgressBar v-if="mem.swap_total > 0" label="Swap" :value="mem.swap_pct"
        :right="`${fmtBytes(mem.swap_used)} / ${fmtBytes(mem.swap_total)}`" />
    </template>

    <TimeChart ref="chart"
      :series="[{ color: '#3fb950', fill: true }]"
      :max="100" :height="72" />
  </section>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { MemStats } from '../types'
import { fmtBytes, pctClass } from '../utils'
import TimeChart from './TimeChart.vue'
import ProgressBar from './ProgressBar.vue'

const props = defineProps<{ mem: MemStats | null }>()
const chart = ref<InstanceType<typeof TimeChart>>()

watch(() => props.mem?.used_pct, v => { if (v != null) chart.value?.push(v) })
</script>

<style scoped>
.panel-hdr   { display:flex; justify-content:space-between; align-items:baseline; }
.panel-title { font-size:11px; font-weight:600; text-transform:uppercase; letter-spacing:.07em; color:var(--muted); }
.panel-sub   { font-size:11px; font-weight:700; }
</style>
