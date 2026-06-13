package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/windskyer/gpu-monitor/internal/alert"
	"github.com/windskyer/gpu-monitor/internal/collector"
	"github.com/windskyer/gpu-monitor/internal/config"
	"github.com/windskyer/gpu-monitor/internal/gpu"
	"github.com/windskyer/gpu-monitor/internal/model"
	"github.com/windskyer/gpu-monitor/internal/notify"
	"github.com/windskyer/gpu-monitor/internal/server"
	"github.com/windskyer/gpu-monitor/internal/store"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	log.Printf("config loaded: listen=%s sample=%s mount_prefix=%q",
		cfg.Server.Listen, cfg.Sample.System, cfg.Host.MountPrefix)

	// GPU collector — non-fatal if no GPU or no driver present
	var gpuColl *gpu.Collector
	if g, err := gpu.NewCollector(); err != nil {
		log.Printf("GPU unavailable: %v", err)
	} else {
		gpuColl = g
		defer gpuColl.Shutdown()
		log.Printf("GPU collector initialized")
	}

	sys := collector.NewSystem(cfg.Host.MountPrefix)
	ring := store.NewRing(cfg.Sample.HistorySz)
	sched := store.NewScheduler(sys, gpuColl, ring, cfg.Sample.System.Duration)

	tg := notify.NewTelegram(cfg.Telegram.BotToken, cfg.Telegram.ChatID)
	alertEngine := alert.NewEngine(cfg.Alerts, tg)
	reporter := notify.NewReporter(tg, ring, cfg.Telegram.ReportInterval.Duration, cfg.Server.Listen, cfg.Server.Domain)

	srv := server.New(ring)

	sched.AddListener(srv.Listener)
	sched.AddListener(alertEngine.Evaluate)

	// Send startup notification after first snapshot so we have real hw info.
	var startupSent bool
	sched.AddListener(func(snap *model.Snapshot) {
		if startupSent {
			return
		}
		startupSent = true
		gpuInfo := "No GPU"
		if len(snap.GPUs) > 0 {
			g := snap.GPUs[0]
			gpuInfo = fmt.Sprintf("%s (Memory %d MB, Temp %d°C)", g.Name, g.MemTotal/1024/1024, g.TempC)
		}
		// gather disk summary (top 2)
		diskStr := "No disk info"
		if len(snap.Disks) > 0 {
			parts := []string{}
			for i, d := range snap.Disks {
				if i >= 2 {
					break
				}
				parts = append(parts, fmt.Sprintf("%s: %.1f%%", d.Mountpoint, d.UsedPct))
			}
			diskStr = strings.Join(parts, "; ")
		}

		msg := fmt.Sprintf("🚀 <b>GPU Monitor started</b>\n"+
			"🖥 Monitoring URL: %s\n"+
			"💻 CPU: %.1f%%  Temp: %.1f°C\n"+
			"🧠 Memory: %.1f%% (%s/%s)\n"+
			"💽 Disks: %s\n"+
			"🎮 GPU: %s",
			cfg.Server.Domain, snap.CPU.UsagePct, snap.CPU.TempC,
			snap.Memory.UsedPct, notify.FmtBytes(snap.Memory.UsedBytes), notify.FmtBytes(snap.Memory.TotalBytes),
			diskStr, gpuInfo)
		if err := tg.Send(msg); err != nil {
			log.Printf("[notify] startup message: %v", err)
		}
	})

	stop := make(chan struct{})
	go sched.Run(stop)
	go reporter.Run(stop)

	httpSrv := &http.Server{
		Addr:    cfg.Server.Listen,
		Handler: srv.Handler(),
	}
	go func() {
		log.Printf("listening on %s", cfg.Server.Listen)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down...")
	close(stop)
}
