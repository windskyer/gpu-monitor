<template>
  <section class="panel">
    <div class="panel-hdr"><span class="panel-title">Disk</span></div>
    <div v-for="d in disks" :key="d.mountpoint" class="disk-item">
      <div class="disk-hdr">
        <span class="disk-mp">{{ d.mountpoint }}</span>
        <span class="disk-io">
          {{ fmtBps(d.read_bps) }} R&nbsp;&nbsp;{{ fmtBps(d.write_bps) }} W
        </span>
      </div>
      <ProgressBar :label="devBasename(d.device)" :value="d.used_pct"
        :right="`${fmtBytes(d.used)} / ${fmtBytes(d.total)}`" />
    </div>
    <span v-if="!disks.length" class="empty">No disks</span>
  </section>
</template>

<script setup lang="ts">
import type { Disk } from '../types';
import { fmtBps, fmtBytes } from '../utils';
import ProgressBar from './ProgressBar.vue';

defineProps<{ disks: Disk[] }>()

function devBasename(dev: string) {
  return dev.split('/').pop() ?? dev
}
</script>

<style scoped>
.panel-title { font-size:11px; font-weight:600; text-transform:uppercase; letter-spacing:.07em; color:var(--muted); }
.disk-item   { margin:4px 0; }
.disk-hdr    { display:flex; justify-content:space-between; font-size:11px; margin-bottom:3px; }
.disk-mp     { font-weight:600; color:var(--text); }
.disk-io     { color:var(--dim); }
.empty       { color:var(--dim); font-size:12px; }
</style>
