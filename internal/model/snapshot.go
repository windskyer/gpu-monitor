package model

import "time"

type Snapshot struct {
	Timestamp time.Time `json:"ts"`
	CPU       CPUStats  `json:"cpu"`
	Memory    MemStats  `json:"mem"`
	Disks     []Disk    `json:"disks"`
	Networks  []NetIface `json:"networks"`
	GPUs      []GPU     `json:"gpus"`
}

type CPUStats struct {
	UsagePct  float64 `json:"usage_pct"`
	TempC     float64 `json:"temp_c"`
	LoadAvg1  float64 `json:"load1"`
	LoadAvg5  float64 `json:"load5"`
	LoadAvg15 float64 `json:"load15"`
	FreqMHz   float64 `json:"freq_mhz"`
}

type MemStats struct {
	TotalBytes uint64  `json:"total"`
	UsedBytes  uint64  `json:"used"`
	FreeBytes  uint64  `json:"free"`
	UsedPct    float64 `json:"used_pct"`
	// Swap
	SwapTotal uint64  `json:"swap_total"`
	SwapUsed  uint64  `json:"swap_used"`
	SwapPct   float64 `json:"swap_pct"`
}

type Disk struct {
	Mountpoint string  `json:"mountpoint"`
	Device     string  `json:"device"`
	Fstype     string  `json:"fstype"`
	TotalBytes uint64  `json:"total"`
	UsedBytes  uint64  `json:"used"`
	FreeBytes  uint64  `json:"free"`
	UsedPct    float64 `json:"used_pct"`
	// IO rates (bytes/s), computed by scheduler from delta
	ReadBps  float64 `json:"read_bps"`
	WriteBps float64 `json:"write_bps"`
}

type NetIface struct {
	Name    string  `json:"name"`
	RecvBps float64 `json:"recv_bps"`
	SendBps float64 `json:"send_bps"`
	// cumulative counters (for rate calc in scheduler)
	RecvBytes uint64 `json:"recv_bytes"`
	SendBytes uint64 `json:"send_bytes"`
}

type GPU struct {
	Index      int    `json:"index"`
	Name       string `json:"name"`
	UUID       string `json:"uuid"`

	// Utilization
	GPUUtilPct uint32 `json:"gpu_util_pct"`
	MemUtilPct uint32 `json:"mem_util_pct"`
	EncUtilPct uint32 `json:"enc_util_pct"`
	DecUtilPct uint32 `json:"dec_util_pct"`

	// Memory (bytes)
	MemUsed  uint64 `json:"mem_used"`
	MemFree  uint64 `json:"mem_free"`
	MemTotal uint64 `json:"mem_total"`

	// Temperature (°C) — only core, no junction (not supported on consumer cards)
	TempC uint32 `json:"temp_c"`

	// Power (W)
	PowerW     float64 `json:"power_w"`
	PowerLimitW float64 `json:"power_limit_w"`

	// Fan (%)
	FanSpeedPct uint32 `json:"fan_speed_pct"`

	// Clocks (MHz)
	ClockGraphicsMHz uint32 `json:"clock_graphics_mhz"`
	ClockSMMHz       uint32 `json:"clock_sm_mhz"`
	ClockMemMHz      uint32 `json:"clock_mem_mhz"`

	// PCIe
	PCIeGen       int     `json:"pcie_gen"`
	PCIeWidth     int     `json:"pcie_width"`
	PCIeTxBps     float64 `json:"pcie_tx_bps"`
	PCIeRxBps     float64 `json:"pcie_rx_bps"`

	// Throttle reasons (bit mask decoded to human string list)
	ThrottleReasons []string `json:"throttle_reasons,omitempty"`
}
