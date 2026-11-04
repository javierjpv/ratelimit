package ratelimiting

import (
	"sync"
	"time"
)

type Limiter struct {
	mu     sync.Mutex
	burst  int
	limit  float64
	tokens float64
	last   time.Time
}

func NewLimiter(b int, l float64) *Limiter {
	return &Limiter{burst: b, limit: l, tokens: float64(b)}
}

func (l *Limiter) Burst() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.burst
}

func (l *Limiter) Limit() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.limit
}

func (l *Limiter) Tokens() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.tokens
}

func (l *Limiter) Last() time.Time {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.last
}
