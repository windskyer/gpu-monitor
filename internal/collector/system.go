package collector

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/sensors"
	"github.com/windskyer/gpu-monitor/internal/model"
)

// SystemCollector gathers CPU/memory/disk/network metrics using gopsutil.
type SystemCollector struct {
	mountPrefix string // e.g. "/host/root" in container, "" on bare metal
}

func NewSystem(mountPrefix string) *SystemCollector {
	return &SystemCollector{mountPrefix: mountPrefix}
}

func (s *SystemCollector) Collect() (*model.Snapshot, error) {
	snap := &model.Snapshot{}

	// CPU usage (blocking ~100ms sample)
	pcts, err := cpu.Percent(0, false)
	if err != nil {
		log.Printf("[collector] cpu percent: %v", err)
	} else if len(pcts) > 0 {
		snap.CPU.UsagePct = math.Round(pcts[0]*100) / 100
	}

	// CPU frequency
	freqs, err := cpu.Info()
	if err == nil && len(freqs) > 0 {
		snap.CPU.FreqMHz = freqs[0].Mhz
	}

	// Load averages
	avg, err := load.Avg()
	if err != nil {
		log.Printf("[collector] load avg: %v", err)
	} else {
		snap.CPU.LoadAvg1 = avg.Load1
		snap.CPU.LoadAvg5 = avg.Load5
		snap.CPU.LoadAvg15 = avg.Load15
	}

	// CPU temperature (k10temp/coretemp via sensors)
	snap.CPU.TempC = cpuTemp()

	// Memory
	vm, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("[collector] virtual memory: %v", err)
	} else {
		snap.Memory = model.MemStats{
			TotalBytes: vm.Total,
			UsedBytes:  vm.Used,
			FreeBytes:  vm.Free,
			UsedPct:    math.Round(vm.UsedPercent*100) / 100,
		}
	}
	sw, err := mem.SwapMemory()
	if err == nil {
		snap.Memory.SwapTotal = sw.Total
		snap.Memory.SwapUsed = sw.Used
		if sw.Total > 0 {
			snap.Memory.SwapPct = math.Round(sw.UsedPercent*100) / 100
		}
	}

	// Disks — enumerate partitions then query usage with prefix
	parts, err := disk.Partitions(false)
	if err != nil {
		log.Printf("[collector] disk partitions: %v", err)
	} else {
		for _, p := range parts {
			// Skip pseudo/virtual filesystems
			if isVirtualFS(p.Fstype) {
				continue
			}
			prefixed := s.mountPrefix + p.Mountpoint
			usage, err := disk.Usage(prefixed)
			if err != nil {
				log.Printf("[collector] disk usage %s: %v", prefixed, err)
				continue
			}
			snap.Disks = append(snap.Disks, model.Disk{
				Mountpoint: p.Mountpoint,
				Device:     p.Device,
				Fstype:     p.Fstype,
				TotalBytes: usage.Total,
				UsedBytes:  usage.Used,
				FreeBytes:  usage.Free,
				UsedPct:    math.Round(usage.UsedPercent*100) / 100,
			})
		}
	}

	// Network — collect cumulative counters; rates are computed by Scheduler
	ifaces, err := net.IOCounters(true)
	if err != nil {
		log.Printf("[collector] net io: %v", err)
	} else {
		for _, iface := range ifaces {
			if iface.Name == "lo" || strings.HasPrefix(iface.Name, "veth") ||
				strings.HasPrefix(iface.Name, "docker") || strings.HasPrefix(iface.Name, "br-") {
				continue
			}
			snap.Networks = append(snap.Networks, model.NetIface{
				Name:      iface.Name,
				RecvBytes: iface.BytesRecv,
				SendBytes: iface.BytesSent,
			})
		}
	}

	return snap, nil
}

// cpuTemp reads the host CPU temperature via gopsutil sensors.
// It looks for k10temp (AMD) or coretemp (Intel) Tdie/Tctl/Package sensors.
func cpuTemp() float64 {
	temps, err := sensors.SensorsTemperatures()
	if err != nil {
		return 0
	}
	// Priority: Tdie > Tctl > Package > first match
	candidates := map[string]float64{}
	for _, t := range temps {
		key := strings.ToLower(t.SensorKey)
		if strings.Contains(key, "tdie") || strings.Contains(key, "tctl") ||
			strings.Contains(key, "package id 0") || strings.Contains(key, "k10temp") {
			candidates[key] = t.Temperature
		}
	}
	for _, k := range []string{"tdie", "tctl", "package id 0"} {
		for key, val := range candidates {
			if strings.Contains(key, k) {
				return val
			}
		}
	}
	// Fallback: first non-zero
	for _, t := range temps {
		if t.Temperature > 0 {
			return t.Temperature
		}
	}
	return 0
}

func isVirtualFS(fstype string) bool {
	virtual := map[string]bool{
		"tmpfs": true, "devtmpfs": true, "sysfs": true, "proc": true,
		"devpts": true, "cgroup": true, "cgroup2": true, "overlay": true,
		"squashfs": true, "nsfs": true, "bpf": true, "pstore": true,
		"mqueue": true, "hugetlbfs": true, "debugfs": true, "tracefs": true,
		"configfs": true, "fusectl": true, "securityfs": true,
	}
	return virtual[fstype]
}

// NetCounters returns a map of interface -> (recv, send) bytes for rate calc.
func NetCounters() (map[string][2]uint64, error) {
	ifaces, err := net.IOCounters(true)
	if err != nil {
		return nil, fmt.Errorf("net.IOCounters: %w", err)
	}
	m := make(map[string][2]uint64, len(ifaces))
	for _, iface := range ifaces {
		m[iface.Name] = [2]uint64{iface.BytesRecv, iface.BytesSent}
	}
	return m, nil
}

// DiskIOCounters returns a map of device -> (read, write) bytes for rate calc.
func DiskIOCounters() (map[string][2]uint64, error) {
	counters, err := disk.IOCounters()
	if err != nil {
		return nil, fmt.Errorf("disk.IOCounters: %w", err)
	}
	m := make(map[string][2]uint64, len(counters))
	for dev, c := range counters {
		m[dev] = [2]uint64{c.ReadBytes, c.WriteBytes}
	}
	return m, nil
}
