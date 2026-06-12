package collector

import "github.com/windskyer/gpu-monitor/internal/model"

// Collector is the common interface for all system-metric sources.
type Collector interface {
	// Collect gathers current metrics. On partial failure the returned snapshot
	// fields are zeroed/nil and the error is logged; the caller should not abort.
	Collect() (*model.Snapshot, error)
}
