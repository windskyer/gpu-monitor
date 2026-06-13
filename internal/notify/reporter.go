package notify

import (
	"fmt"
	"log"
	"time"

	"strings"

	"github.com/windskyer/gpu-monitor/internal/model"
	"github.com/windskyer/gpu-monitor/internal/store"
)

// Reporter sends periodic summary reports to Telegram.
type Reporter struct {
	tg       *Telegram
	ring     *store.Ring
	interval time.Duration
	listen   string
}

func NewReporter(tg *Telegram, ring *store.Ring, interval time.Duration, listen string) *Reporter {
	return &Reporter{tg: tg, ring: ring, interval: interval, listen: listen}
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
	body := formatReport(snap)
	msg := body
	if r.listen != "" {
		msg = fmt.Sprintf("🚀 <b>GPU Monitor</b>\n🖥 监控地址: http://%s\n\n%s", r.listen, body)
	}
	if err := r.tg.Send(msg); err != nil {
		log.Printf("[reporter] send: %v", err)
	}
}

func formatReport(s *model.Snapshot) string {
	t := s.Timestamp.Format("2006-01-02 15:04:05")
	txt := fmt.Sprintf("<b>📊 GPU Monitor Report</b> — %s\n\n", t)

	// include server listen if available (reporter will prefix when called)
	// CPU
	txt += fmt.Sprintf("<b>CPU</b> %.1f%% | %.1f°C | Freq %.2f GHz | Load %.2f %.2f %.2f\n",
		s.CPU.UsagePct, s.CPU.TempC, s.CPU.FreqMHz/1000.0, s.CPU.LoadAvg1, s.CPU.LoadAvg5, s.CPU.LoadAvg15)

	// Memory
	txt += fmt.Sprintf("<b>Memory</b> %.1f%% (%s / %s)\n",
		s.Memory.UsedPct, fmtBytes(s.Memory.UsedBytes), fmtBytes(s.Memory.TotalBytes))

	// Disks
	if len(s.Disks) > 0 {
		txt += "\n<b>Disks</b>\n"
		for _, d := range s.Disks {
			txt += fmt.Sprintf("- %s: %.1f%% (%s/%s)\n", d.Mountpoint, d.UsedPct, fmtBytes(d.UsedBytes), fmtBytes(d.TotalBytes))
		}
	}

	// Networks
	if len(s.Networks) > 0 {
		txt += "\n<b>Network</b>\n"
		for _, n := range s.Networks {
			txt += fmt.Sprintf("- %s: ↓ %s ↑ %s\n", n.Name, fmtBps(n.RecvBps), fmtBps(n.SendBps))
		}
	}

	// GPUs
	if len(s.GPUs) > 0 {
		txt += "\n<b>GPUs</b>\n"
		for _, g := range s.GPUs {
			memPct := 0.0
			if g.MemTotal > 0 {
				memPct = float64(g.MemUsed) / float64(g.MemTotal) * 100
			}
			// escape name to avoid accidental HTML
			name := escHTML(g.Name)
			txt += fmt.Sprintf("- %s: Util %d%% | Mem %.1f%% (%s/%s) | Temp %d°C | %.0fW\n",
				name, g.GPUUtilPct, memPct, fmtBytes(g.MemUsed), fmtBytes(g.MemTotal), g.TempC, g.PowerW)
		}
	}

	return txt
}

// escHTML escapes < and > to avoid injecting raw tags in parse_mode=HTML
func escHTML(s string) string {
	out := s
	out = strings.ReplaceAll(out, "<", "&lt;")
	out = strings.ReplaceAll(out, ">", "&gt;")
	return out
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

func fmtBps(b float64) string {
	if b < 0 {
		b = 0
	}
	switch {
	case b >= 1e9:
		return fmt.Sprintf("%.2f GB/s", b/1e9)
	case b >= 1e6:
		return fmt.Sprintf("%.1f MB/s", b/1e6)
	case b >= 1e3:
		return fmt.Sprintf("%.0f KB/s", b/1e3)
	default:
		return fmt.Sprintf("%.0f B/s", b)
	}
}
