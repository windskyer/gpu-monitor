package alert

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/windskyer/gpu-monitor/internal/config"
	"github.com/windskyer/gpu-monitor/internal/model"
)

type state int

const (
	stateIdle    state = iota
	stateFiring
)

type ruleState struct {
	s         state
	lastFired time.Time
}

// Notifier sends an alert message.
type Notifier interface {
	Send(msg string) error
}

// Engine evaluates alert rules and fires notifications.
type Engine struct {
	cfg      config.AlertsCfg
	notifier Notifier

	mu     sync.Mutex
	states map[string]*ruleState
}

func NewEngine(cfg config.AlertsCfg, n Notifier) *Engine {
	return &Engine{cfg: cfg, notifier: n, states: make(map[string]*ruleState)}
}

func (e *Engine) Evaluate(snap *model.Snapshot) {
	now := snap.Timestamp

	// CPU temperature
	e.check("cpu_temp", snap.CPU.TempC >= e.cfg.CPUTemp.Threshold,
		fmt.Sprintf("CPU temperature %.1f°C >= %.0f°C", snap.CPU.TempC, e.cfg.CPUTemp.Threshold),
		e.cfg.CPUTemp.Cooldown, now)

	// Memory %
	e.check("mem_pct", snap.Memory.UsedPct >= e.cfg.MemPct.Threshold,
		fmt.Sprintf("Memory usage %.1f%% >= %.0f%%", snap.Memory.UsedPct, e.cfg.MemPct.Threshold),
		e.cfg.MemPct.Cooldown, now)

	// Disk %
	for _, d := range snap.Disks {
		key := "disk_pct:" + d.Mountpoint
		e.check(key, d.UsedPct >= e.cfg.DiskPct.Threshold,
			fmt.Sprintf("Disk %s usage %.1f%% >= %.0f%%", d.Mountpoint, d.UsedPct, e.cfg.DiskPct.Threshold),
			e.cfg.DiskPct.Cooldown, now)
	}

	// GPU alerts
	for _, g := range snap.GPUs {
		prefix := fmt.Sprintf("gpu%d", g.Index)

		e.check(prefix+":temp", float64(g.TempC) >= e.cfg.GPUTemp.Threshold,
			fmt.Sprintf("GPU %d (%s) temperature %d°C >= %.0f°C", g.Index, g.Name, g.TempC, e.cfg.GPUTemp.Threshold),
			e.cfg.GPUTemp.Cooldown, now)

		var memPct float64
		if g.MemTotal > 0 {
			memPct = float64(g.MemUsed) / float64(g.MemTotal) * 100
		}
		e.check(prefix+":mem", memPct >= e.cfg.GPUMemPct.Threshold,
			fmt.Sprintf("GPU %d (%s) memory %.1f%% >= %.0f%%", g.Index, g.Name, memPct, e.cfg.GPUMemPct.Threshold),
			e.cfg.GPUMemPct.Cooldown, now)
	}
}

func (e *Engine) check(key string, firing bool, msg string, cooldown time.Duration, now time.Time) {
	e.mu.Lock()
	rs, ok := e.states[key]
	if !ok {
		rs = &ruleState{}
		e.states[key] = rs
	}
	e.mu.Unlock()

	if !firing {
		rs.s = stateIdle
		return
	}
	if rs.s == stateFiring && now.Sub(rs.lastFired) < cooldown {
		return
	}
	rs.s = stateFiring
	rs.lastFired = now
	go func() {
		text := "⚠️ ALERT: " + msg
		if err := e.notifier.Send(text); err != nil {
			log.Printf("[alert] send %q: %v", key, err)
		}
	}()
}
