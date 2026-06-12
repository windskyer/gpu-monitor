// Package store provides the ring buffer and the Scheduler that drives
// periodic collection and fan-out to all consumers.
package store

import (
	"log"
	"sync"
	"time"

	"github.com/windskyer/gpu-monitor/internal/collector"
	"github.com/windskyer/gpu-monitor/internal/gpu"
	"github.com/windskyer/gpu-monitor/internal/model"
)

// Listener receives each new snapshot.
type Listener func(*model.Snapshot)

// Scheduler drives periodic collection and fan-outs.
type Scheduler struct {
	sys      *collector.SystemCollector
	gpuColl  *gpu.Collector // nil if no GPU
	ring     *Ring
	interval time.Duration

	mu        sync.RWMutex
	listeners []Listener

	// Rate-calculation state
	prevNetBytes  map[string][2]uint64 // iface -> [recv, send]
	prevDiskBytes map[string][2]uint64 // device -> [read, write]
	prevTime      time.Time
}

func NewScheduler(sys *collector.SystemCollector, g *gpu.Collector, ring *Ring, interval time.Duration) *Scheduler {
	return &Scheduler{
		sys:      sys,
		gpuColl:  g,
		ring:     ring,
		interval: interval,
	}
}

func (s *Scheduler) AddListener(l Listener) {
	s.mu.Lock()
	s.listeners = append(s.listeners, l)
	s.mu.Unlock()
}

func (s *Scheduler) Run(stop <-chan struct{}) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			snap := s.collect()
			if snap == nil {
				continue
			}
			s.ring.Push(snap)
			s.mu.RLock()
			ls := make([]Listener, len(s.listeners))
			copy(ls, s.listeners)
			s.mu.RUnlock()
			for _, l := range ls {
				l(snap)
			}
		}
	}
}

func (s *Scheduler) collect() *model.Snapshot {
	snap, err := s.sys.Collect()
	if err != nil {
		log.Printf("[scheduler] system collect: %v", err)
		snap = &model.Snapshot{}
	}
	snap.Timestamp = time.Now()

	// GPU
	if s.gpuColl != nil {
		gpus, err := s.gpuColl.Collect()
		if err != nil {
			log.Printf("[scheduler] gpu collect: %v", err)
		} else {
			snap.GPUs = gpus
		}
	}

	// Rate calculations
	now := snap.Timestamp
	if !s.prevTime.IsZero() {
		dt := now.Sub(s.prevTime).Seconds()
		if dt > 0 {
			s.applyNetRates(snap, dt)
			s.applyDiskRates(snap, dt)
		}
	}

	// Save network counters for next tick (disk prev state is managed inside applyDiskRates)
	s.prevNetBytes = netCountersFromSnap(snap)
	s.prevTime = now

	return snap
}

func (s *Scheduler) applyNetRates(snap *model.Snapshot, dt float64) {
	for i := range snap.Networks {
		iface := &snap.Networks[i]
		if prev, ok := s.prevNetBytes[iface.Name]; ok {
			iface.RecvBps = float64(iface.RecvBytes-prev[0]) / dt
			iface.SendBps = float64(iface.SendBytes-prev[1]) / dt
		}
	}
}

func (s *Scheduler) applyDiskRates(snap *model.Snapshot, dt float64) {
	diskIO, err := collector.DiskIOCounters()
	if err != nil {
		return
	}
	if s.prevDiskBytes == nil {
		s.prevDiskBytes = make(map[string][2]uint64, len(diskIO))
	}
	// Update prev for all devices so next tick has a baseline
	for dev, cur := range diskIO {
		if prev, ok := s.prevDiskBytes[dev]; ok {
			// Find the matching disk in snap and set rates
			for i := range snap.Disks {
				if deviceBasename(snap.Disks[i].Device) == dev {
					snap.Disks[i].ReadBps = float64(cur[0]-prev[0]) / dt
					snap.Disks[i].WriteBps = float64(cur[1]-prev[1]) / dt
				}
			}
		}
		s.prevDiskBytes[dev] = cur
	}
}

func netCountersFromSnap(snap *model.Snapshot) map[string][2]uint64 {
	m := make(map[string][2]uint64, len(snap.Networks))
	for _, n := range snap.Networks {
		m[n.Name] = [2]uint64{n.RecvBytes, n.SendBytes}
	}
	return m
}

func deviceBasename(dev string) string {
	for i := len(dev) - 1; i >= 0; i-- {
		if dev[i] == '/' {
			return dev[i+1:]
		}
	}
	return dev
}
