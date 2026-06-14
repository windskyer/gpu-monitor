<template>
  <section
    class="panel"
    :class="{ clickable: Boolean(link) }"
    :tabindex="link ? 0 : undefined"
    :role="link ? 'link' : undefined"
    @click="openLink"
    @keydown.enter="openLink"
    @keydown.space.prevent="openLink">
    <div class="panel-hdr"><span class="panel-title">Network</span></div>
    <div v-for="iface in nets" :key="iface.name" class="iface">
      <div class="iface-hdr">
        <span class="iface-name">{{ iface.name }}</span>
        <div class="iface-rates">
          <span class="rx">↓ {{ fmtBps(iface.recv_bps) }}</span>
          <span class="tx">↑ {{ fmtBps(iface.send_bps) }}</span>
        </div>
      </div>
      <TimeChart :ref="el => setChart(iface.name, el as ChartInst)"
        :series="[{ color: '#58a6ff', fill: true }, { color: '#bc8cff', fill: false }]"
        :height="40" />
    </div>
    <span v-if="!nets.length" class="empty">No interfaces</span>
  </section>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import type { NetIface } from '../types'
import { fmtBps } from '../utils'
import TimeChart from './TimeChart.vue'

type ChartInst = InstanceType<typeof TimeChart>

const props = defineProps<{
  nets: NetIface[]
  link?: string
}>()

const chartMap = new Map<string, ChartInst>()
function setChart(name: string, el: ChartInst | null) {
  if (el) chartMap.set(name, el)
}

function openLink() {
  if (!props.link) return
  location.assign(props.link)
}

watch(() => props.nets, (nets) => {
  nets.forEach(n => chartMap.get(n.name)?.push(n.recv_bps, n.send_bps))
}, { deep: true })
</script>

<style scoped>
.panel-title { font-size:11px; font-weight:600; text-transform:uppercase; letter-spacing:.07em; color:var(--muted); }
.clickable   { cursor:pointer; transition:border-color .16s, background .16s; }
.clickable:hover,
.clickable:focus-visible { border-color:var(--blue); background:var(--raised); outline:none; }
.iface       { border-bottom:1px solid var(--border); padding:6px 0; }
.iface:last-child { border-bottom:none; }
.iface-hdr   { display:flex; justify-content:space-between; align-items:baseline; margin-bottom:4px; }
.iface-name  { font-weight:600; color:var(--blue); }
.iface-rates { display:flex; gap:10px; font-size:11px; }
.rx { color:var(--green); }
.tx { color:var(--purple); }
.empty { color:var(--dim); font-size:12px; }
</style>
