export interface Snapshot {
  ts:       string
  cpu:      CPUStats
  mem:      MemStats
  disks:    Disk[]    | null
  networks: NetIface[] | null
  gpus:     GPU[]     | null
}

export interface CPUStats {
  usage_pct: number
  temp_c:    number
  load1:     number
  load5:     number
  load15:    number
  freq_mhz:  number
}

export interface MemStats {
  total:      number
  used:       number
  free:       number
  used_pct:   number
  swap_total: number
  swap_used:  number
  swap_pct:   number
}

export interface Disk {
  mountpoint: string
  device:     string
  fstype:     string
  total:      number
  used:       number
  free:       number
  used_pct:   number
  read_bps:   number
  write_bps:  number
}

export interface NetIface {
  name:       string
  recv_bps:   number
  send_bps:   number
  recv_bytes: number
  send_bytes: number
}

export interface GPU {
  index:             number
  name:              string
  uuid:              string
  gpu_util_pct:      number
  mem_util_pct:      number
  enc_util_pct:      number
  dec_util_pct:      number
  mem_used:          number
  mem_free:          number
  mem_total:         number
  temp_c:            number
  power_w:           number
  power_limit_w:     number
  fan_speed_pct:     number
  clock_graphics_mhz: number
  clock_sm_mhz:      number
  clock_mem_mhz:     number
  pcie_gen:          number
  pcie_width:        number
  pcie_tx_bps:       number
  pcie_rx_bps:       number
  throttle_reasons:  string[] | null
}
