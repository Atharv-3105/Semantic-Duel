package game

import "time"

type State struct {
	Status Status
	EndsAt time.Time
}

type Status string

const (
	Waiting  Status = "WAITING"
	Active   Status = "ACTIVE"
	Finished Status = "FINISHED"
)

// StartGame transitions the state to Active and sets the end time.
// durationSeconds comes from config (GAME_DURATION_SECONDS env var).
func StartGame(state *State, durationSeconds int) {
	state.Status = Active
	state.EndsAt = time.Now().Add(time.Duration(durationSeconds) * time.Second)
}

func IsGameOver(state *State) bool {
	return time.Now().After(state.EndsAt)
}