package ratelimiting

import (
	"context"
	"errors"
	"fmt"
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
	return &Limiter{limit: l, burst: b, tokens: float64(b)}, nil
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
func (l *Limiter) futureTokens(date time.Time) float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	// calculates the elapsed time between last update on tokens field and the date to be compared
	duration := date.Sub(l.last)
	tokens := duration.Seconds()*l.limit + l.tokens
	return tokens
}

// t is the amount of tokens to be reserved
func (l *Limiter) reserve(t float64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	currentTokens := l.futureTokens(time.Now())
	remaining := currentTokens - t
	l.tokens = remaining
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
	if t > float64(l.burst) {
		return errors.New("tokens can not be greater than burst")
	}
	select {
	case <-ctx.Done():
		return nil
	//default allows the function to continue without unrecheable code
	default:
	}
	currentTokens := l.futureTokens(time.Now())
	remaining := currentTokens - t
	fmt.Println("remaining", remaining)
	if remaining >= 0 {
		l.reserve(t)
	} else {
		waitDuration, err := l.timeTowait(remaining)
		fmt.Println("waitDuration", waitDuration)
		if err != nil {
			return err
		}
		timer := time.NewTimer(waitDuration)
		select {
		// go routine will be sleeping here until 1 case triggers
		// when a case triggers, function can continue to be executed
		case <-ctx.Done():
			return nil
		case <-timer.C:
		}
	}
	return nil
}
