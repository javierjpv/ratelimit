package ratelimiting

import (
	"errors"
	"sync"
	"time"
)

type tokens float64

type Limiter struct {
	mu     sync.Mutex
	burst  int
	limit  float64
	tokens float64
	last   time.Time
}

func NewLimiter(l float64, b int) *Limiter {
	return &Limiter{limit: l, burst: b, tokens: float64(b)}
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



// t is the amount of tokens to be consumed
// calculates the elapsed time between last update on tokens field and the date to be compared
// beware of context bleeding
func (l *Limiter) generatedTokens(date time.Time) float64 {
	duration := date.Sub(l.last)

	tokens := duration.Seconds() * l.limit
	return tokens
}

//t is the amount of tokens to be reserved
func (l *Limiter) reserveN(t float64) float64 {
	result := l.generatedTokens(time.Now()) - t
	return result
}

//t is the amount of pending tokens to be re generated
func (l *Limiter) timeTowait(t float64) (tim time.Duration, err error) {
	if t <= 0 {
		return 0, errors.New("it can not be 0 or minor")
	}
	durationF := t / float64(l.limit)

	duration := (durationF * float64(time.Second))

	return time.Duration(duration), nil
}
