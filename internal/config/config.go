package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerCfg   `yaml:"server"`
	Host     HostCfg     `yaml:"host"`
	Sample   SampleCfg   `yaml:"sample"`
	Telegram TelegramCfg `yaml:"telegram"`
	Alerts   AlertsCfg   `yaml:"alerts"`
	Exporter ExporterCfg `yaml:"exporter"`
}

type ServerCfg struct {
	Listen string `yaml:"listen"`
	Token  string `yaml:"token"`
}

type HostCfg struct {
	MountPrefix string `yaml:"mount_prefix"`
}

type SampleCfg struct {
	System    time.Duration `yaml:"system"`
	GPU       time.Duration `yaml:"gpu"`
	Process   time.Duration `yaml:"process"`
	HistorySz int           `yaml:"history"`
	TopN      int           `yaml:"top_n"`
}

type TelegramCfg struct {
	BotToken       string        `yaml:"bot_token"`
	ChatID         string        `yaml:"chat_id"`
	ReportInterval time.Duration `yaml:"report_interval"`
}

type AlertsCfg struct {
	GPUTemp    AlertRule `yaml:"gpu_temp"`
	GPUMemPct  AlertRule `yaml:"gpu_mem_pct"`
	CPUTemp    AlertRule `yaml:"cpu_temp"`
	MemPct     AlertRule `yaml:"mem_pct"`
	DiskPct    AlertRule `yaml:"disk_pct"`
}

type AlertRule struct {
	Threshold float64       `yaml:"threshold"`
	Cooldown  time.Duration `yaml:"cooldown"`
}

type ExporterCfg struct {
	Prometheus bool `yaml:"prometheus"`
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg := defaults()
	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func defaults() *Config {
	return &Config{
		Server: ServerCfg{Listen: "0.0.0.0:8800", Token: "change-me"},
		Sample: SampleCfg{
			System:    time.Second,
			GPU:       time.Second,
			Process:   2 * time.Second,
			HistorySz: 180,
			TopN:      15,
		},
		Telegram: TelegramCfg{ReportInterval: time.Hour},
		Alerts: AlertsCfg{
			GPUTemp:   AlertRule{Threshold: 83, Cooldown: 10 * time.Minute},
			GPUMemPct: AlertRule{Threshold: 95, Cooldown: 10 * time.Minute},
			CPUTemp:   AlertRule{Threshold: 85, Cooldown: 10 * time.Minute},
			MemPct:    AlertRule{Threshold: 90, Cooldown: 10 * time.Minute},
			DiskPct:   AlertRule{Threshold: 90, Cooldown: 30 * time.Minute},
		},
	}
}
