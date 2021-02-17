package middleware

import (
	"sync"
	"time"
)

type metaData struct {
	count int
	reset time.Time
}

type memLimiter struct {
	mu       sync.Mutex
	ipLimit  int
	duration time.Duration
	table    map[string]metaData
}

func newMemLimiter(opts *Options) *Limiter {
	mem := &memLimiter{
		ipLimit:  opts.IPLimit,
		duration: opts.Duration,
		table:    make(map[string]metaData),
	}
	return &Limiter{mem}
}

func (m *memLimiter) getLimit(key string, policy ...int) ([]interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var md metaData
	if data, ok := m.table[key]; !ok || !data.reset.After(time.Now()) {
		md = metaData{
			count: m.ipLimit - 1,
			reset: time.Now().Add(time.Duration((m.duration / time.Millisecond)) * time.Millisecond), // round time delta to milisec
		}
		m.table[key] = data
	} else {
		md = metaData{
			count: data.count - 1,
			reset: data.reset,
		}
		if md.count < -1 {
			md.count = -1
		}
	}
	m.table[key] = md
	return []interface{}{md.count, md.reset}, nil
}

func (m *memLimiter) removeLimit(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.table, key)
	return nil
}
