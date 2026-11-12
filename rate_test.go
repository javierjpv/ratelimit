package ratelimiting_test

import (
	"testing"
	ratelimiting "github.com/javierjpv/ratelimit"
)

func TestLimiter(t *testing.T) {

	tests := []struct {
		name       string
		limit      float64
		burst      int
		wantNil    bool
		wantBurst  int
		wantLimit  float64
		wantTokens float64
	}{
		{
			name:       "first",
			limit:      2,
			burst:      4,
			wantNil:    false,
			wantBurst:  4,
			wantLimit:  2,
			wantTokens: 4,
		},
	}

	for _, tt := range tests {
		gotLimiter := ratelimiting.NewLimiter(tt.limit, tt.burst)
		if tt.wantNil && gotLimiter != nil {
			t.Fatal("not expected wantNil")
		}
		if !tt.wantNil && gotLimiter == nil {
			t.Fatal("not expected wantNil")
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
