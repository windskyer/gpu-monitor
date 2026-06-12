<template>
  <section class="panel gpu-panel">
    <!-- Header -->
    <div class="gpu-hdr">
      <div class="gpu-name"><span class="live-dot" />&nbsp;GPU {{ g.index }} &mdash; {{ g.name }}</div>
      <div class="badges">
        <span class="badge"><span class="bl">Temp</span>
          <span :class="tempClass(g.temp_c)">{{ g.temp_c }}°C</span></span>
        <span class="badge"><span class="bl">Power</span>
          <span>{{ g.power_w.toFixed(0) }}<span class="bl">/{{ g.power_limit_w.toFixed(0) }}W</span></span></span>
        <span class="badge"><span class="bl">Fan</span> {{ g.fan_speed_pct }}%</span>
        <span class="badge"><span class="bl">PCIe</span> Gen{{ g.pcie_gen }}×{{ g.pcie_width }}</span>
      </div>
    </div>

    <!-- Metric chips -->
    <div class="chip-row">
      <div class="chip">
        <span class="chip-lbl">GPU Util</span>
        <span class="chip-val" :class="pctClass(g.gpu_util_pct)">{{ g.gpu_util_pct }}%</span>
      </div>
      <div class="chip">
        <span class="chip-lbl">VRAM</span>
        <span class="chip-val c-cyan">{{ fmtBytes(g.mem_used) }}</span>
        <span class="chip-sub">/ {{ fmtBytes(g.mem_total) }}</span>
      </div>
      <div class="chip">
        <span class="chip-lbl">VRAM %</span>
        <span class="chip-val" :class="pctClass(memPct)">{{ memPct.toFixed(1) }}%</span>
      </div>
      <div class="chip">
        <span class="chip-lbl">Enc</span>
        <span class="chip-val c-purple">{{ g.enc_util_pct }}%</span>
      </div>
      <div class="chip">
        <span class="chip-lbl">Dec</span>
        <span class="chip-val c-purple">{{ g.dec_util_pct }}%</span>
      </div>
      <div class="chip">
        <span class="chip-lbl">Graphics</span>
        <span class="chip-val">{{ g.clock_graphics_mhz }}<span class="chip-sub"> MHz</span></span>
      </div>
    </div>

    <!-- Charts -->
    <div class="chart-area">
      <div class="chart-wrap">
        <div class="chart-lbl">
          <span>GPU Utilization</span>
          <span class="cv">{{ g.gpu_util_pct }}%</span>
        </div>
        <TimeChart ref="utilChart"
          :series="[{ color: '#bc8cff', fill: true }]"
          :max="100" :height="90" />
      </div>
      <div class="chart-wrap">
        <div class="chart-lbl">
          <span>VRAM</span>
          <span class="cv">{{ fmtBytes(g.mem_used) }} / {{ fmtBytes(g.mem_total) }}</span>
        </div>
        <TimeChart ref="memChart"
          :series="[{ color: '#39d353', fill: true }]"
          :max="g.mem_total" :height="90" />
      </div>
    </div>

    <!-- Footer: clocks · PCIe · throttle -->
    <div class="gpu-footer">
      <div class="fgrp">
        <span class="fl">GFX</span> <span class="fv">{{ g.clock_graphics_mhz }} MHz</span>
        <span class="fl">SM</span>  <span class="fv">{{ g.clock_sm_mhz }} MHz</span>
        <span class="fl">MEM</span> <span class="fv">{{ g.clock_mem_mhz }} MHz</span>
      </div>
      <div class="fgrp">
        <span class="fl">TX</span> <span class="fv">{{ fmtBps(g.pcie_tx_bps) }}</span>
        <span class="fl">RX</span> <span class="fv">{{ fmtBps(g.pcie_rx_bps) }}</span>
      </div>
      <div v-if="g.throttle_reasons && g.throttle_reasons.length" class="fgrp">
        <span class="fl">Throttle</span>
        <span v-for="r in g.throttle_reasons" :key="r" class="tag">{{ r }}</span>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { GPU } from '../types'
import { fmtBytes, fmtBps, tempClass, pctClass } from '../utils'
import TimeChart from './TimeChart.vue'

const props = defineProps<{ g: GPU }>()

const memPct = computed(() =>
  props.g.mem_total > 0 ? (props.g.mem_used / props.g.mem_total) * 100 : 0)

const utilChart = ref<InstanceType<typeof TimeChart>>()
const memChart  = ref<InstanceType<typeof TimeChart>>()

watch(() => props.g, (g) => {
  utilChart.value?.push(g.gpu_util_pct)
  memChart.value?.push(g.mem_used)
}, { deep: false })
</script>

<style scoped>
.gpu-panel { display:flex; flex-direction:column; gap:8px; }

.gpu-hdr  { display:flex; align-items:center; justify-content:space-between; flex-wrap:wrap; gap:8px; }
.gpu-name { font-size:13px; font-weight:700; display:flex; align-items:center; gap:6px; }
.live-dot { width:7px; height:7px; border-radius:50%; background:var(--green); animation:blink 2.5s ease-in-out infinite; flex-shrink:0; }
@keyframes blink { 0%,100%{opacity:1} 50%{opacity:.4} }

.badges { display:flex; gap:6px; flex-wrap:wrap; }
.badge  { display:inline-flex; align-items:center; gap:4px; padding:2px 8px; border-radius:99px;
          font-size:11px; font-weight:600; background:var(--raised); border:1px solid var(--border2); white-space:nowrap; }
.bl { color:var(--dim); font-weight:400; margin-right:2px; }

.chip-row { display:flex; gap:6px; flex-wrap:wrap; }
.chip     { background:var(--raised); border:1px solid var(--border); border-radius:var(--r);
            padding:5px 10px; display:flex; flex-direction:column; align-items:center; gap:1px; min-width:76px; flex:1; }
.chip-lbl { font-size:10px; color:var(--muted); text-transform:uppercase; letter-spacing:.06em; }
.chip-val { font-size:18px; font-weight:700; line-height:1.1; }
.chip-sub { font-size:9px; color:var(--dim); }

.chart-area { display:grid; grid-template-columns:2fr 1fr; gap:8px; }
.chart-wrap { display:flex; flex-direction:column; gap:3px; }
.chart-lbl  { display:flex; justify-content:space-between; align-items:baseline; font-size:10px; color:var(--dim); }
.chart-lbl .cv { font-size:11px; font-weight:600; color:var(--muted); }

.gpu-footer { display:flex; flex-wrap:wrap; gap:14px; font-size:11px; color:var(--muted);
              border-top:1px solid var(--border); padding-top:6px; }
.fgrp { display:flex; gap:8px; align-items:center; flex-wrap:wrap; }
.fl   { color:var(--dim); }
.fv   { color:var(--text); }
.tag  { background:rgba(248,81,73,.12); color:var(--red); border:1px solid rgba(248,81,73,.3);
        border-radius:3px; padding:1px 5px; font-size:10px; }
</style>
