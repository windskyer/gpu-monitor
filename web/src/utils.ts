export function fmtBytes(b: number): string {
  if (b >= 1099511627776) return (b / 1099511627776).toFixed(1) + ' TB'
  if (b >= 1073741824)    return (b / 1073741824).toFixed(1) + ' GB'
  if (b >= 1048576)       return (b / 1048576).toFixed(1) + ' MB'
  if (b >= 1024)          return (b / 1024).toFixed(1) + ' KB'
  return b + ' B'
}

export function fmtBps(bps: number): string {
  if (bps < 0) bps = 0
  if (bps >= 1e9) return (bps / 1e9).toFixed(2) + ' GB/s'
  if (bps >= 1e6) return (bps / 1e6).toFixed(1) + ' MB/s'
  if (bps >= 1e3) return (bps / 1e3).toFixed(0) + ' KB/s'
  return bps.toFixed(0) + ' B/s'
}

/** Round up to a "nice" number for Y-axis auto-scaling. */
export function niceMax(v: number): number {
  if (v <= 0) return 1
  const e = Math.pow(10, Math.floor(Math.log10(v)))
  return ([1, 2, 5, 10].map(f => f * e).find(n => n >= v)) ?? v
}

export function tempClass(c: number): string {
  if (c >= 90) return 'c-red'
  if (c >= 80) return 'c-orange'
  if (c >= 70) return 'c-yellow'
  if (c >= 60) return 'c-cyan'
  return 'c-green'
}

export function pctClass(p: number): string {
  if (p >= 90) return 'c-red'
  if (p >= 70) return 'c-orange'
  if (p >= 50) return 'c-yellow'
  return 'c-green'
}

export function barColor(p: number): string {
  if (p >= 90) return 'var(--red)'
  if (p >= 70) return 'var(--orange)'
  if (p >= 50) return 'var(--yellow)'
  return 'var(--green)'
}

// Canvas palette — keeps JS in sync with CSS vars
export const C = {
  bg:     '#0d1117',
  grid:   'rgba(33,38,45,0.85)',
  dim:    '#484f58',
  green:  '#3fb950',
  cyan:   '#39d353',
  blue:   '#58a6ff',
  purple: '#bc8cff',
  yellow: '#d29922',
  orange: '#f0883e',
  red:    '#f85149',
} as const
