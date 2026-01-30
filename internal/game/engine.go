package game

import (
	"time"
)

type State struct {
	Status Status 
	EndsAt time.Time
}

type Status string 

const (
	Waiting Status = "WAITING"
	Active Status  = "ACTIVE"
	Finished Status = "FINISHED"
)

func StartGame(state *State) {
	state.Status = Active
	state.EndsAt = time.Now().Add(60 * time.Second)
}


func IsGameOver(state *State) bool {
	return time.Now().After(state.EndsAt)
}