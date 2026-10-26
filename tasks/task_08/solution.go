package main

import (
	"math"
	"sync"
	"time"
)

type Clock interface{ Now() time.Time }
type Limiter struct {
	mu     sync.Mutex
	clock  Clock
	rate   float64
	burst  int
	tokens float64
	last   time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	l := &Limiter{
		clock: clock,
		rate:  ratePerSec,
		burst: burst,
	}

	if clock == nil || burst <= 0 {
		return l
	}

	l.tokens = float64(burst)
	l.last = clock.Now()

	return l
}
func (l *Limiter) AllowN(n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if n <= 0 || n > l.burst {
		return false
	}
	if l.burst <= 0 {
		return false
	}
	if l.clock == nil {
		return false
	}

	now := l.clock.Now()
	if now.After(l.last) {
		time_passed := now.Sub(l.last).Seconds()

		if l.rate > 0 {
			l.tokens = math.Min(
				float64(l.burst),
				l.tokens+time_passed*l.rate,
			)
		}

		l.last = now
	}

	if l.tokens < float64(n) {
		return false
	}

	l.tokens -= float64(n)
	return true
}
