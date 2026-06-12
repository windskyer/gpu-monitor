package notify

import (
	"fmt"
	"log"
	"time"

	"github.com/windskyer/gpu-monitor/internal/model"
	"github.com/windskyer/gpu-monitor/internal/store"
)

// Reporter sends periodic summary reports to Telegram.
type Reporter struct {
	tg       *Telegram
	ring     *store.Ring
	interval time.Duration
}

func NewReporter(tg *Telegram, ring *store.Ring, interval time.Duration) *Reporter {
	return &Reporter{tg: tg, ring: ring, interval: interval}
}

func (r *Reporter) Run(stop <-chan struct{}) {
	if r.interval <= 0 {
		return
	}
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			r.send()
		}
	}
}

func (r *Reporter) send() {
	snap := r.ring.Latest()
	if snap == nil {
		return
	}
	msg := formatReport(snap)
	if err := r.tg.Send(msg); err != nil {
		log.Printf("[reporter] send: %v", err)
	}
}

func formatReport(s *model.Snapshot) string {
	t := s.Timestamp.Format("2006-01-02 15:04:05")
	txt := fmt.Sprintf("<b>📊 GPU Monitor Report</b> — %s\n\n", t)

	txt += fmt.Sprintf("<b>CPU</b> %.1f%% | %.1f°C | Load %.2f %.2f %.2f\n",
		s.CPU.UsagePct, s.CPU.TempC, s.CPU.LoadAvg1, s.CPU.LoadAvg5, s.CPU.LoadAvg15)

	txt += fmt.Sprintf("<b>Memory</b> %.1f%% (%s / %s)\n",
		s.Memory.UsedPct, fmtBytes(s.Memory.UsedBytes), fmtBytes(s.Memory.TotalBytes))

	for _, g := range s.GPUs {
		memPct := 0.0
		if g.MemTotal > 0 {
			memPct = float64(g.MemUsed) / float64(g.MemTotal) * 100
		}
		txt += fmt.Sprintf("\n<b>GPU %d</b> %s\n", g.Index, g.Name)
		txt += fmt.Sprintf("  Util %d%% | Mem %.1f%% (%s/%s) | Temp %d°C | %.0fW\n",
			g.GPUUtilPct, memPct,
			fmtBytes(g.MemUsed), fmtBytes(g.MemTotal),
			g.TempC, g.PowerW)
	}

	for _, d := range s.Disks {
		txt += fmt.Sprintf("<b>Disk</b> %s %.1f%% (%s/%s)\n",
			d.Mountpoint, d.UsedPct, fmtBytes(d.UsedBytes), fmtBytes(d.TotalBytes))
	}
	return txt
}

func fmtBytes(b uint64) string {
	const (
		GB = 1 << 30
		MB = 1 << 20
		KB = 1 << 10
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.1fG", float64(b)/GB)
	case b >= MB:
		return fmt.Sprintf("%.1fM", float64(b)/MB)
	case b >= KB:
		return fmt.Sprintf("%.1fK", float64(b)/KB)
	default:
		return fmt.Sprintf("%dB", b)
	}
}

// FmtBytes is exported for server use.
func FmtBytes(b uint64) string { return fmtBytes(b) }
