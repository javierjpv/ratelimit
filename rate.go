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

