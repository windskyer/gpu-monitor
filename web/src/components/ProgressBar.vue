<template>
  <div class="bar-row">
    <span class="bar-lbl" :title="label">{{ label }}</span>
    <div class="bar-track">
      <div class="bar-fill" :style="{ width: pct + '%', background: color }" />
    </div>
    <span class="bar-val">{{ right }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { barColor } from '../utils'

const props = defineProps<{
  label: string
  value: number   // 0–100
  right: string   // right-side text
  color?: string
}>()

const pct   = computed(() => Math.min(Math.max(props.value, 0), 100))
const color = computed(() => props.color ?? barColor(pct.value))
</script>

<style scoped>
.bar-row  { display:flex; align-items:center; gap:6px; margin:3px 0; }
.bar-lbl  { font-size:11px; color:var(--muted); min-width:60px; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; flex-shrink:0; }
.bar-track{ flex:1; height:6px; background:var(--bg); border-radius:3px; overflow:hidden; }
.bar-fill { height:100%; border-radius:3px; transition:width .5s ease, background .5s ease; }
.bar-val  { font-size:10px; color:var(--muted); min-width:56px; text-align:right; white-space:nowrap; flex-shrink:0; }
</style>
