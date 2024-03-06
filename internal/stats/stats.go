package stats

import (
	"sync"

	"github.com/anchore/go-sync/internal/atomic"
)

type Stat uint32

// Stats provides some additional statistics for each node
type Stats interface {
	// Add adds the value to the given stat line
	Add(stat Stat, delta float64)

	// Set sets the value for the given stat line
	Set(stat Stat, value float64)

	// Get gets the value of the given stat line
	Get(stat Stat) (value float64)
}

// ---------- statsSlice implementation ----------

type statsSlice []atomic.Float64

var _ Stats = (*statsSlice)(nil)

func NewStats(stats ...Stat) Stats {
	maxStatNum := -1
	for _, s := range stats {
		if int(s) > maxStatNum {
			maxStatNum = int(s)
		}
	}
	if maxStatNum < 0 {
		panic("no stats provided")
	}
	s := make(statsSlice, maxStatNum+1)
	return &s
}

func (s *statsSlice) Add(stat Stat, value float64) {
	if s == nil || int(stat) >= len(*s) {
		return
	}
	(&(*s)[stat]).Add(value)
}

func (s *statsSlice) Set(stat Stat, value float64) {
	if s == nil || int(stat) >= len(*s) {
		return
	}
	(&(*s)[stat]).Store(value)
}

func (s *statsSlice) Get(stat Stat) (value float64) {
	if s == nil || int(stat) >= len(*s) {
		return 0
	}
	return (&(*s)[stat]).Load()
}

// --- statsMap ---

type statsMap struct {
	sync.RWMutex
	stats map[Stat]float64
}

var _ Stats = (*statsMap)(nil)

func (s *statsMap) Add(stat Stat, value float64) {
	s.Lock()
	defer s.Unlock()
	if s.stats == nil {
		s.stats = map[Stat]float64{}
	}
	s.stats[stat] += value
}

func (s *statsMap) Set(stat Stat, value float64) {
	s.Lock()
	defer s.Unlock()
	if s.stats == nil {
		s.stats = map[Stat]float64{}
	}
	s.stats[stat] = value
}

func (s *statsMap) Get(stat Stat) (value float64) {
	s.RLock()
	defer s.RUnlock()
	if s.stats == nil {
		return 0
	}
	return s.stats[stat]
}

// --- statsAtomic ---

type statsAtomic struct {
	sync.RWMutex
	stats map[Stat]*atomic.Float64
}

var _ Stats = (*statsAtomic)(nil)

func (s *statsAtomic) _get(stat Stat) *atomic.Float64 {
	s.RLock()
	rUnlock := s.RUnlock
	if s.stats == nil {
		rUnlock()
		s.Lock()
		unlock := s.Unlock
		if s.stats == nil {
			s.stats = map[Stat]*atomic.Float64{}
		}
		unlock()
		s.RLock()
		rUnlock = s.RUnlock
	}

	v := s.stats[stat]
	if v == nil {
		rUnlock()
		s.Lock()
		unlock := s.Unlock
		v = s.stats[stat]
		if v == nil {
			v = &atomic.Float64{}
			s.stats[stat] = v
		}
		unlock()
	} else {
		rUnlock()
	}
	return v
}

func (s *statsAtomic) Add(stat Stat, delta float64) {
	s._get(stat).Add(delta)
}

func (s *statsAtomic) Set(stat Stat, value float64) {
	s._get(stat).Store(value)
}

func (s *statsAtomic) Get(stat Stat) (value float64) {
	s.RLock()
	defer s.RUnlock()
	if s.stats == nil {
		return 0
	}
	v := s.stats[stat]
	if v != nil {
		return v.Load()
	}
	return 0
}
