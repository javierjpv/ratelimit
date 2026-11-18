package ratelimiting

import (
	"context"
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

func NewLimiter(l float64, b int) (*Limiter, error) {
	if l <= 0 {
		return nil, errors.New("limit can not be 0 or negative")
	}
	if b <= 0 {
		return nil, errors.New("burst can not be 0 or negative")
	}
	return &Limiter{limit: l, burst: b, tokens: float64(b), last: time.Now()}, nil
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
// beware of context bleeding
// futureTokens requires l.mu to be held by the caller.
func (l *Limiter) futureTokens(date time.Time) float64 {
	// calculates the elapsed time between last update on tokens field and the date to be compared
	duration := date.Sub(l.last)
	tokens := duration.Seconds()*l.limit + l.tokens
	if tokens > float64(l.burst) {
		tokens = float64(l.burst)
	}
	return tokens
}

// t is the amount of tokens to be reserved
// updateTokens requires l.mu to be held by the caller.
func (l *Limiter) updateTokens(t float64, currentTokens float64, currentTime time.Time) {
	l.tokens = currentTokens - t
	l.last = currentTime
}

// t is the amount of pending tokens to be regenerated
func (l *Limiter) timeTowait(t float64) (tim time.Duration, err error) {
	if t <= 0 {
		return 0, errors.New("it can not be 0 or negative")
	}
	durationF := t / float64(l.limit)
	duration := (durationF * float64(time.Second))
	return time.Duration(duration), nil
}

func (l *Limiter) Wait(ctx context.Context) error {
	err := l.wait(ctx, 1)
	return err
}

// t is the number of tokens to be consumed
func (l *Limiter) WaitN(ctx context.Context, t float64) error {
	err := l.wait(ctx, t)
	return err
}

// t is the amount of tokens to be consumed
// sleep the go routine depending of the select condition
func (l *Limiter) wait(ctx context.Context, t float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	//default allows the function to continue without unrecheable code
	default:
	}
	l.mu.Lock()
	if t > float64(l.burst) {
		l.mu.Unlock()
		return errors.New("tokens can not be greater than burst")
	}

	now := time.Now()
	currentTokens := l.futureTokens(now)
	remaining := currentTokens - t
	l.updateTokens(t, currentTokens, now)

	if remaining >= 0 {
		l.mu.Unlock()
		return nil
	}
	waitDuration, err := l.timeTowait(-remaining)
	l.mu.Unlock()
	if err != nil {
		return err
	}
	timer := time.NewTimer(waitDuration)
	defer timer.Stop()
	select {
	// go routine will be sleeping here until 1 case triggers
	// when a case triggers, function can continue to be executed
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
	}

	return nil
}

func (l *Limiter) allow(t float64) (bool, error) {
	if t > float64(l.burst) {
		return false, errors.New("limit can not be greater than burst")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	result := l.futureTokens(now) - t
	if result >= 0 {
		l.tokens = result
		l.last = now
		return true, nil
	}
	return false, nil
}
func (l *Limiter) Allow() (bool, error) {
	return l.allow(1)
}

func (l *Limiter) AllowN(t float64) (bool, error) {
	return l.allow(t)
}

type Reservation struct {
	duration time.Duration
}

func (l *Limiter) reserve(t float64) (Reservation, error) {
	l.mu.Lock()
	if t > float64(l.burst) {
		l.mu.Unlock()
		return Reservation{duration: 0}, errors.New("tokens can not be greater than burst")
	}

	now := time.Now()
	currentTokens := l.futureTokens(now)
	remaining := currentTokens - t
	l.updateTokens(t, currentTokens, now)

	if remaining >= 0 {
		l.mu.Unlock()
		return Reservation{duration: 0}, nil
	}
	waitDuration, err := l.timeTowait(-remaining)
	l.mu.Unlock()
	if err != nil {
		return Reservation{duration: 0}, err
	}
	return Reservation{duration: waitDuration}, nil
}

func (l *Limiter) Reserve(ctx context.Context) (Reservation, error) {
	select {
	case <-ctx.Done():
		return Reservation{duration: 0}, ctx.Err()
	default:

	}
	reservation, err := l.reserve(1)
	return reservation, err
}

func (l *Limiter) ReserveN(ctx context.Context, t float64) (Reservation, error) {
	select {
	case <-ctx.Done():
		return Reservation{duration: 0}, ctx.Err()
	default:
	}
	reservation, err := l.reserve(t)
	return reservation, err
}
