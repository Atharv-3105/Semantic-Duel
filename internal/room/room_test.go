package room

import (
	"testing"
)


func TestRoomKeepsBestScore(t *testing.T){
	r := &Room{
		State:	&GameState{
			Scores: make(map[string]int),
		},
	}

	r.UpdateScore("p1", 0.6)
	r.UpdateScore("p1", 0.4)
	r.UpdateScore("p1", 0.9)

	if r.State.Scores["p1"] != 90 {
		t.Fatalf("expected best score 90, got %d", r.State.Scores["p1"])
	}
}