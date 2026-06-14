package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration wraps time.Duration so yaml can decode human strings like "1s", "10m".
type Duration struct{ time.Duration }

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	v := value.Value
	dur, err := time.ParseDuration(v)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", v, err)
	}
	d.Duration = dur
	return nil
}

type Config struct {
	Server   ServerCfg   `yaml:"server"`
	Host     HostCfg     `yaml:"host"`
	Sample   SampleCfg   `yaml:"sample"`
	Telegram TelegramCfg `yaml:"telegram"`
	Alerts   AlertsCfg   `yaml:"alerts"`
	Exporter ExporterCfg `yaml:"exporter"`
}

type ServerCfg struct {
	Listen     string `yaml:"listen"`
	Domain     string `yaml:"domain"`
	Token      string `yaml:"token"`
	NetworkURL string `yaml:"network_url"`
}

type HostCfg struct {
	MountPrefix string `yaml:"mount_prefix"`
}

type SampleCfg struct {
	System    Duration `yaml:"system"`
	GPU       Duration `yaml:"gpu"`
	Process   Duration `yaml:"process"`
	HistorySz int      `yaml:"history"`
	TopN      int      `yaml:"top_n"`
}

type TelegramCfg struct {
	BotToken       string   `yaml:"bot_token"`
	ChatID         string   `yaml:"chat_id"`
	ReportInterval Duration `yaml:"report_interval"`
}

type AlertsCfg struct {
	GPUTemp   AlertRule `yaml:"gpu_temp"`
	GPUMemPct AlertRule `yaml:"gpu_mem_pct"`
	CPUTemp   AlertRule `yaml:"cpu_temp"`
	MemPct    AlertRule `yaml:"mem_pct"`
	DiskPct   AlertRule `yaml:"disk_pct"`
}

type AlertRule struct {
	Threshold float64  `yaml:"threshold"`
	Cooldown  Duration `yaml:"cooldown"`
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
		Server: ServerCfg{Listen: "0.0.0.0:8800"},
		Sample: SampleCfg{
			System:    Duration{time.Second},
			GPU:       Duration{time.Second},
			Process:   Duration{2 * time.Second},
			HistorySz: 180,
			TopN:      15,
		},
		Telegram: TelegramCfg{ReportInterval: Duration{time.Hour}},
		Alerts: AlertsCfg{
			GPUTemp:   AlertRule{Threshold: 83, Cooldown: Duration{10 * time.Minute}},
			GPUMemPct: AlertRule{Threshold: 95, Cooldown: Duration{10 * time.Minute}},
			CPUTemp:   AlertRule{Threshold: 85, Cooldown: Duration{10 * time.Minute}},
			MemPct:    AlertRule{Threshold: 90, Cooldown: Duration{10 * time.Minute}},
			DiskPct:   AlertRule{Threshold: 90, Cooldown: Duration{30 * time.Minute}},
		},
	}
}
