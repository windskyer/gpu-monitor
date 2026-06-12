package gpu

import (
	"fmt"
	"log"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/windskyer/gpu-monitor/internal/model"
)

// Collector wraps go-nvml and produces GPU snapshots.
type Collector struct{}

func NewCollector() (*Collector, error) {
	if ret := nvml.Init(); ret != nvml.SUCCESS {
		return nil, fmt.Errorf("nvml.Init: %v", nvml.ErrorString(ret))
	}
	return &Collector{}, nil
}

func (c *Collector) Shutdown() {
	nvml.Shutdown()
}

// Collect returns one model.GPU per device. Partial field failures are logged
// and zeroed — the caller always gets a best-effort result.
func (c *Collector) Collect() ([]model.GPU, error) {
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("DeviceGetCount: %v", nvml.ErrorString(ret))
	}

	gpus := make([]model.GPU, 0, count)
	for i := 0; i < count; i++ {
		dev, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			log.Printf("[gpu] DeviceGetHandleByIndex(%d): %v", i, nvml.ErrorString(ret))
			continue
		}
		g := collectOne(i, dev)
		gpus = append(gpus, g)
	}
	return gpus, nil
}

func collectOne(idx int, dev nvml.Device) model.GPU {
	g := model.GPU{Index: idx}

	if name, ret := dev.GetName(); ret == nvml.SUCCESS {
		g.Name = name
	}
	if uuid, ret := dev.GetUUID(); ret == nvml.SUCCESS {
		g.UUID = uuid
	}

	// Utilization
	if u, ret := dev.GetUtilizationRates(); ret == nvml.SUCCESS {
		g.GPUUtilPct = u.Gpu
		g.MemUtilPct = u.Memory
	} else {
		log.Printf("[gpu] %d GetUtilizationRates: %v", idx, nvml.ErrorString(ret))
	}

	// Encoder / Decoder
	if enc, _, ret := dev.GetEncoderUtilization(); ret == nvml.SUCCESS {
		g.EncUtilPct = enc
	}
	if dec, _, ret := dev.GetDecoderUtilization(); ret == nvml.SUCCESS {
		g.DecUtilPct = dec
	}

	// Memory
	if mi, ret := dev.GetMemoryInfo(); ret == nvml.SUCCESS {
		g.MemUsed = mi.Used
		g.MemFree = mi.Free
		g.MemTotal = mi.Total
	} else {
		log.Printf("[gpu] %d GetMemoryInfo: %v", idx, nvml.ErrorString(ret))
	}

	// Temperature — core only; FI_DEV_MEMORY_TEMP not supported on consumer cards
	if temp, ret := dev.GetTemperature(nvml.TEMPERATURE_GPU); ret == nvml.SUCCESS {
		g.TempC = temp
	} else {
		log.Printf("[gpu] %d GetTemperature: %v", idx, nvml.ErrorString(ret))
	}

	// Power (mW → W)
	if pw, ret := dev.GetPowerUsage(); ret == nvml.SUCCESS {
		g.PowerW = float64(pw) / 1000.0
	}
	if pl, ret := dev.GetEnforcedPowerLimit(); ret == nvml.SUCCESS {
		g.PowerLimitW = float64(pl) / 1000.0
	}

	// Fan
	if fan, ret := dev.GetFanSpeed(); ret == nvml.SUCCESS {
		g.FanSpeedPct = fan
	}

	// Clocks
	if clk, ret := dev.GetClockInfo(nvml.CLOCK_GRAPHICS); ret == nvml.SUCCESS {
		g.ClockGraphicsMHz = clk
	}
	if clk, ret := dev.GetClockInfo(nvml.CLOCK_SM); ret == nvml.SUCCESS {
		g.ClockSMMHz = clk
	}
	if clk, ret := dev.GetClockInfo(nvml.CLOCK_MEM); ret == nvml.SUCCESS {
		g.ClockMemMHz = clk
	}

	// PCIe
	if gen, ret := dev.GetCurrPcieLinkGeneration(); ret == nvml.SUCCESS {
		g.PCIeGen = gen
	}
	if width, ret := dev.GetCurrPcieLinkWidth(); ret == nvml.SUCCESS {
		g.PCIeWidth = width
	}
	// PCIe throughput in KB/s from NVML — store raw here; scheduler converts to bps
	if tx, ret := dev.GetPcieThroughput(nvml.PCIE_UTIL_TX_BYTES); ret == nvml.SUCCESS {
		g.PCIeTxBps = float64(tx) * 1024
	}
	if rx, ret := dev.GetPcieThroughput(nvml.PCIE_UTIL_RX_BYTES); ret == nvml.SUCCESS {
		g.PCIeRxBps = float64(rx) * 1024
	}

	// Throttle reasons — use EventReasons (NVML 13+), not deprecated ThrottleReasons
	if reasons, ret := dev.GetCurrentClocksEventReasons(); ret == nvml.SUCCESS {
		g.ThrottleReasons = decodeThrottleReasons(reasons)
	}

	return g
}

// Throttle reason bit masks (from nvml.h)
const (
	reasonGPUIdle               uint64 = 0x0000000000000001
	reasonApplicationsClocksSetting uint64 = 0x0000000000000002
	reasonSwPowerCap            uint64 = 0x0000000000000004
	reasonHWSlowdown            uint64 = 0x0000000000000008
	reasonSyncBoost             uint64 = 0x0000000000000010
	reasonSwThermalSlowdown     uint64 = 0x0000000000000020
	reasonHWThermalSlowdown     uint64 = 0x0000000000000040
	reasonHWPowerBrakeSlowdown  uint64 = 0x0000000000000080
	reasonDisplayClockSetting   uint64 = 0x0000000000000100
)

func decodeThrottleReasons(mask uint64) []string {
	if mask == 0 || mask == reasonGPUIdle {
		return nil
	}
	var out []string
	if mask&reasonApplicationsClocksSetting != 0 {
		out = append(out, "app_clocks")
	}
	if mask&reasonSwPowerCap != 0 {
		out = append(out, "sw_power_cap")
	}
	if mask&reasonHWSlowdown != 0 {
		out = append(out, "hw_slowdown")
	}
	if mask&reasonSyncBoost != 0 {
		out = append(out, "sync_boost")
	}
	if mask&reasonSwThermalSlowdown != 0 {
		out = append(out, "sw_thermal")
	}
	if mask&reasonHWThermalSlowdown != 0 {
		out = append(out, "hw_thermal")
	}
	if mask&reasonHWPowerBrakeSlowdown != 0 {
		out = append(out, "hw_power_brake")
	}
	if mask&reasonDisplayClockSetting != 0 {
		out = append(out, "display_clocks")
	}
	return out
}
