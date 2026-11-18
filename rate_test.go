package ratelimiting_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	ratelimiting "github.com/javierjpv/ratelimit"
)

func TestLimiter(t *testing.T) {

	tests := []struct {
		name       string
		limit      float64
		burst      int
		wantBurst  int
		wantLimit  float64
		wantTokens float64
		wantError  bool
	}{
		{
			name:       "first",
			limit:      2,
			burst:      4,
			wantBurst:  4,
			wantLimit:  2,
			wantTokens: 4,
			wantError:  false,
		},
		{
			name:       "second",
			limit:      0,
			burst:      0,
			wantBurst:  0,
			wantLimit:  0,
			wantTokens: 0,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		gotLimiter, err := ratelimiting.NewLimiter(tt.limit, tt.burst)

		if err != nil && !tt.wantError {
			t.Fatal("not expected wantError")
		}
		if err != nil && tt.wantError {
			continue
		}
		if gotLimiter.Limit() != tt.wantLimit {
			t.Fatal("not expected wantLimit")
		}
		if gotLimiter.Burst() != tt.wantBurst {
			t.Fatal("not expected wantBurst")
		}
		if gotLimiter.Tokens() != tt.wantTokens {
			t.Fatal("not expected tokens")
		}

	}

}

func TestWaitNConcurrent(t *testing.T) {

	tests := []struct {
		name      string
		limit     float64
		burst     int
		wantError bool
	}{
		{
			name:      "first",
			limit:     2,
			burst:     4,
			wantError: false,
		},
		{
			name:      "second",
			limit:     2,
			burst:     6,
			wantError: false,
		},
	}

	for _, tt := range tests {
		//avoiding data race
		currentTest := tt
		synctest.Test(t, func(t *testing.T) {
			start := time.Now()
			gotLimiter, err := ratelimiting.NewLimiter(currentTest.limit, currentTest.burst)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var wg sync.WaitGroup
			ctx, _ := context.WithCancel(context.Background())
			totalRequest := 13
			for i := 0; i < totalRequest; i++ {
				go func() {
					wg.Add(1)
					defer wg.Done()
					err := gotLimiter.WaitN(ctx, 1)
					if err != nil && !currentTest.wantError {
						t.Errorf("unexpected error in WaitN: %v", err)
					}
				}()

			}
			//synctest.Wait() allows the test to continue only if the go routine
			//is sleeping waiting for a chan in select
			synctest.Wait()
			wg.Wait()

			// virtual time inside synctest bubble
			elapsed := time.Since(start)

			expectedSec := float64(totalRequest-currentTest.burst) / currentTest.limit
			expected := time.Duration(expectedSec * float64(time.Second))

			if elapsed != expected {
				t.Fatalf("[%s] elapsed time => %v, expected time => %v", currentTest.name, elapsed, expected)
			}
		})

	}
}

func TestWaitNContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	limiter, err := ratelimiting.NewLimiter(2, 4)
	if err != nil {
		t.Fatalf("unexpected error creating limiter: %v", err)
	}

	err = limiter.WaitN(ctx, 1)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected error at WaitN %v", err)
	}
}

func TestAllowN(t *testing.T) {
	tests := []struct {
		name          string
		limit         float64
		burst         int
		wantError     bool
		requestTokens float64
	}{
		{
			name:          "cfirst",
			limit:         2,
			burst:         4,
			wantError:     false,
			requestTokens: 1,
		},
		{
			name:          "second",
			limit:         2,
			burst:         4,
			requestTokens: 20,
			wantError:     true,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			lim, err := ratelimiting.NewLimiter(tt.limit, tt.burst)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			_, err = lim.AllowN(tt.requestTokens)
			if !tt.wantError && err != nil {
				t.Fatalf("unexpected error in AllowN: %v", err)
			}

		})
	}
}

func TestAllowNConcurrent(t *testing.T) {
	tests := []struct {
		name      string
		limit     float64
		burst     int
		wantError bool
	}{
		{
			name:      "first",
			limit:     2,
			burst:     4,
			wantError: false,
		},
		{
			name:      "second",
			limit:     2,
			burst:     6,
			wantError: false,
		},
	}

	for _, tt := range tests {
		//avoiding data race
		currentTest := tt
		t.Run(currentTest.name, func(t *testing.T) {
			lim, err := ratelimiting.NewLimiter(currentTest.limit, currentTest.burst)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var wg sync.WaitGroup
			totalRequests := 40
			for i := 0; i < totalRequests; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, err := lim.AllowN(1)
					if !currentTest.wantError && err != nil {
						t.Errorf("unexpected error in AllowN: %v", err)
					}
				}()

			}
			wg.Wait()
		})
	}
}
