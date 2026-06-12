package store

import (
	"sync"

	"github.com/windskyer/gpu-monitor/internal/model"
)

// Ring is a fixed-capacity circular buffer of Snapshots.
type Ring struct {
	mu   sync.RWMutex
	buf  []*model.Snapshot
	cap  int
	head int // index of next write position
	size int // number of valid entries
}

func NewRing(capacity int) *Ring {
	return &Ring{buf: make([]*model.Snapshot, capacity), cap: capacity}
}

func (r *Ring) Push(s *model.Snapshot) {
	r.mu.Lock()
	r.buf[r.head] = s
	r.head = (r.head + 1) % r.cap
	if r.size < r.cap {
		r.size++
	}
	r.mu.Unlock()
}

// Latest returns the most-recently pushed snapshot, or nil if empty.
func (r *Ring) Latest() *model.Snapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.size == 0 {
		return nil
	}
	idx := (r.head - 1 + r.cap) % r.cap
	return r.buf[idx]
}

// All returns snapshots in chronological order (oldest first).
func (r *Ring) All() []*model.Snapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.size == 0 {
		return nil
	}
	out := make([]*model.Snapshot, r.size)
	start := (r.head - r.size + r.cap) % r.cap
	for i := 0; i < r.size; i++ {
		out[i] = r.buf[(start+i)%r.cap]
	}
	return out
}
