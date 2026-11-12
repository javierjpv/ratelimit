package ratelimiting_test

import (
	ratelimiting "github.com/javierjpv/ratelimit"
	"testing"
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
