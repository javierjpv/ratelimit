package ratelimiting_test

import (
	"context"
	"fmt"
	"testing"
	"testing/synctest"

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

func TestWaitN(t *testing.T) {

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
		synctest.Test(t, func(t *testing.T) {
			gotLimiter, err := ratelimiting.NewLimiter(tt.limit, tt.burst)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			for i := 0; i < 13; i++ {
				go func() {
					err := gotLimiter.WaitN(ctx, 1)
					if err != nil && !tt.wantError {
						t.Fatalf("unexpected error in WaitN: %v", err)
					}
				}()

			}
			//synctest.Wait() allows the test to continue only if the go routine 
			//is sleeping waiting for a chan in select
			synctest.Wait()
			fmt.Println("El test continua tras tener la go routine dormida esperando que legue un chan en el select")

		})

	}
}
