package game

import (
	"testing"
)

func TestSimilarityToScore(t *testing.T) {
	test := []struct{
		sim float64
		want int 
	}{
		{0.0, 0},
		{0.25, 25},
		{0.5, 50},
		{0.82, 82},
		{1.0, 100},
		{1.2, 100},  // clamp
		{-0.5, 0},   // clamp
	}

	for _, tt := range test {
		got  := SimilarityToScore(tt.sim)
		if got != tt.want {
			t.Errorf("SimilarityToScore(%f) = %d, want %d", tt.sim, got, tt.want)
		}
	}
}	