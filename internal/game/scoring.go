package game 

import "math"

func SimilarityToScore(sim float64) int {
	if sim < 0 {
		sim = 0
	}
	if sim > 1 {
		sim = 1
	}
	return int(math.Round(sim * 100))
}