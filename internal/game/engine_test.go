package game

import (
	"testing"
	"time"
)

func TestStartGame_SetsActiveStatus(t *testing.T) {
	state := State{}
	StartGame(&state, 60)

	if state.Status != Active {
		t.Errorf("expected Status=Active, got %s", state.Status)
	}
}

func TestStartGame_RespectsDurationSeconds(t *testing.T) {
	cases := []struct {
		name     string
		duration int
	}{
		{"60 seconds", 60},
		{"90 seconds", 90},
		{"120 seconds", 120},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := State{}
			before := time.Now()
			StartGame(&state, tc.duration)

			expectedEnd := before.Add(time.Duration(tc.duration) * time.Second)

			// Allow a 200ms window for test execution overhead.
			diff := state.EndsAt.Sub(expectedEnd)
			if diff < -200*time.Millisecond || diff > 200*time.Millisecond {
				t.Errorf("EndsAt off by %v for duration %ds", diff, tc.duration)
			}
		})
	}
}

func TestStartGame_OverwritesPreviousState(t *testing.T) {
	state := State{Status: Finished}
	StartGame(&state, 60)

	if state.Status != Active {
		t.Errorf("StartGame should overwrite Finished state, got %s", state.Status)
	}
}

func TestIsGameOver_FalseWhenJustStarted(t *testing.T) {
	state := State{}
	StartGame(&state, 60)

	if IsGameOver(&state) {
		t.Error("game should not be over immediately after starting")
	}
}

func TestIsGameOver_TrueWhenEndsAtIsInPast(t *testing.T) {
	state := State{
		Status: Active,
		EndsAt: time.Now().Add(-1 * time.Second), // expired 1s ago
	}

	if !IsGameOver(&state) {
		t.Error("game should be over when EndsAt is in the past")
	}
}

func TestIsGameOver_FalseWhenEndsAtIsInFuture(t *testing.T) {
	state := State{
		Status: Active,
		EndsAt: time.Now().Add(10 * time.Second),
	}

	if IsGameOver(&state) {
		t.Error("game should not be over when EndsAt is in the future")
	}
}
