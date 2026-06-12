<template>
  <canvas ref="cv" class="chart-cv" :style="{ height: height + 'px' }" />
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { C, niceMax } from '../utils'

interface SeriesDef {
  color: string
  fill?: boolean   // filled area; default true for first series
}

const props = withDefaults(defineProps<{
  series:     SeriesDef[]
  max?:       number       // fixed Y-max; omit for auto-scale
  height?:    number       // CSS height in px
  historyLen?: number
}>(), {
  height:     72,
  historyLen: 120,
})

// ── internal buffers (one per series) ────────────────────────────────────────
const buffers: number[][] = props.series.map(() => [])

const cv = ref<HTMLCanvasElement>()
let observer: ResizeObserver | null = null

// ── public API ────────────────────────────────────────────────────────────────
/** Push one data-point per series, then redraw. */
function push(...values: (number | undefined)[]) {
  values.forEach((v, i) => {
    if (i >= buffers.length) return
    buffers[i].push(v ?? 0)
    if (buffers[i].length > props.historyLen) buffers[i].shift()
  })
  draw()
}

defineExpose({ push })

// ── drawing ───────────────────────────────────────────────────────────────────
function draw() {
  const el  = cv.value
  if (!el) return
  const ctx = el.getContext('2d')!
  const dpr = window.devicePixelRatio || 1
  const W   = el.offsetWidth
  const H   = el.offsetHeight
  if (!W || !H) return

  if (el.width !== Math.round(W * dpr) || el.height !== Math.round(H * dpr)) {
    el.width  = Math.round(W * dpr)
    el.height = Math.round(H * dpr)
  }
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0)

  // background
  ctx.fillStyle = C.bg
  ctx.fillRect(0, 0, W, H)

  const main = buffers[0]
  if (main.length < 2) return

  // Y scale
  let yMax = props.max
  if (yMax == null) {
    const mx = Math.max(...buffers.flat(), 1)
    yMax = niceMax(mx)
  }

  const PR = 36, PT = 4, PB = 2
  const CW = W - PR, CH = H - PT - PB
  const HIST = props.historyLen
  const dx = CW / (HIST - 1)

  // grid lines at 25 / 50 / 75 %
  ctx.strokeStyle = C.grid
  ctx.lineWidth = 1
  for (const f of [0.25, 0.5, 0.75]) {
    const y = PT + CH * (1 - f) + 0.5
    ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(CW, y); ctx.stroke()
  }

  // draw each series
  buffers.forEach((buf, si) => {
    if (buf.length < 2) return
    const def  = props.series[si]
    const col  = def.color
    const fill = def.fill !== false && si === 0
    const off  = HIST - buf.length

    if (fill) {
      ctx.beginPath()
      ctx.moveTo(off * dx, PT + CH)
      for (let i = 0; i < buf.length; i++) {
        ctx.lineTo((off + i) * dx, PT + CH - (buf[i] / yMax!) * CH)
      }
      ctx.lineTo((off + buf.length - 1) * dx, PT + CH)
      ctx.closePath()
      const g = ctx.createLinearGradient(0, PT, 0, PT + CH)
      g.addColorStop(0, col + '55')
      g.addColorStop(1, col + '08')
      ctx.fillStyle = g
      ctx.fill()
    }

    ctx.beginPath()
    for (let i = 0; i < buf.length; i++) {
      const x = (off + i) * dx
      const y = PT + CH - (buf[i] / yMax!) * CH
      i === 0 ? ctx.moveTo(x, y) : ctx.lineTo(x, y)
    }
    ctx.strokeStyle = col
    ctx.lineWidth   = 1.5
    ctx.lineJoin    = 'round'
    ctx.stroke()
  })

  // Y-axis labels (right side)
  ctx.font          = '10px monospace'
  ctx.textAlign     = 'right'
  ctx.textBaseline  = 'middle'
  ctx.fillStyle     = C.dim
  const fmt = (v: number) => props.max != null
    ? Math.round(v / yMax! * 100) + '%'
    : fmtAxisVal(v)
  ctx.fillText(fmt(yMax), W - 2, PT + 5)
  ctx.fillText(fmt(yMax * 0.5), W - 2, PT + CH * 0.5)
}

function fmtAxisVal(v: number): string {
  if (v >= 1e9) return (v / 1e9).toFixed(0) + 'G'
  if (v >= 1e6) return (v / 1e6).toFixed(0) + 'M'
  if (v >= 1e3) return (v / 1e3).toFixed(0) + 'K'
  return v.toFixed(0)
}

onMounted(() => {
  observer = new ResizeObserver(draw)
  observer.observe(cv.value!)
})
onUnmounted(() => observer?.disconnect())
</script>

<style scoped>
.chart-cv { width: 100%; display: block; border-radius: 4px; }
</style>
