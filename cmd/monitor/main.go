package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/windskyer/gpu-monitor/internal/alert"
	"github.com/windskyer/gpu-monitor/internal/collector"
	"github.com/windskyer/gpu-monitor/internal/config"
	"github.com/windskyer/gpu-monitor/internal/gpu"
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

	// GPU collector — non-fatal if no GPU or no driver present
	var gpuColl *gpu.Collector
	if g, err := gpu.NewCollector(); err != nil {
		log.Printf("GPU unavailable: %v", err)
	} else {
		gpuColl = g
		defer gpuColl.Shutdown()
	}

	sys := collector.NewSystem(cfg.Host.MountPrefix)
	ring := store.NewRing(cfg.Sample.HistorySz)
	sched := store.NewScheduler(sys, gpuColl, ring, cfg.Sample.System)

	tg := notify.NewTelegram(cfg.Telegram.BotToken, cfg.Telegram.ChatID)
	alertEngine := alert.NewEngine(cfg.Alerts, tg)
	reporter := notify.NewReporter(tg, ring, cfg.Telegram.ReportInterval)

	srv := server.New(ring, cfg.Server.Token)

	sched.AddListener(srv.Listener)
	sched.AddListener(alertEngine.Evaluate)

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
