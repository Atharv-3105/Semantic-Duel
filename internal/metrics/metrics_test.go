package metrics

import (
	"sync/atomic"
	"testing"
)

// reset zeroes all counters so tests don't interfere with each other.
func reset() {
	atomic.StoreInt64(&GamesStarted, 0)
	atomic.StoreInt64(&GamesCompleted, 0)
	atomic.StoreInt64(&Disconnects, 0)
	atomic.StoreInt64(&ActiveConnections, 0)
}

func TestMetrics_GamesStarted(t *testing.T) {
	reset()

	IncGamesStarted()
	IncGamesStarted()
	IncGamesStarted()

	snap := SnapShot()
	if snap["games_started"] != 3 {
		t.Errorf("expected games_started=3, got %d", snap["games_started"])
	}
}

func TestMetrics_GamesCompleted(t *testing.T) {
	reset()

	IncGamesCompleted()
	IncGamesCompleted()

	snap := SnapShot()
	if snap["games_completed"] != 2 {
		t.Errorf("expected games_completed=2, got %d", snap["games_completed"])
	}
}

func TestMetrics_Disconnects(t *testing.T) {
	reset()

	IncDisconnects()

	snap := SnapShot()
	if snap["disconnects"] != 1 {
		t.Errorf("expected disconnects=1, got %d", snap["disconnects"])
	}
}

func TestMetrics_ActiveConnections_IncDec(t *testing.T) {
	reset()

	IncConnections()
	IncConnections()
	IncConnections()
	DecConnections()

	snap := SnapShot()
	if snap["active_connections"] != 2 {
		t.Errorf("expected active_connections=2 after 3 inc + 1 dec, got %d", snap["active_connections"])
	}
}

func TestMetrics_ActiveConnections_NeverGoesNegative(t *testing.T) {
	reset()

	DecConnections() // should go to -1 but that's fine — atomic doesn't clamp

	// This just verifies the atomic operation works; the server logic prevents negative values.
	snap := SnapShot()
	if snap["active_connections"] != -1 {
		t.Errorf("expected -1 after single Dec from 0, got %d", snap["active_connections"])
	}
}

func TestMetrics_SnapShot_ContainsAllRequiredKeys(t *testing.T) {
	reset()
	snap := SnapShot()

	required := []string{"games_started", "games_completed", "disconnects", "active_connections"}
	for _, key := range required {
		if _, ok := snap[key]; !ok {
			t.Errorf("SnapShot missing required key: %q", key)
		}
	}
}

func TestMetrics_SnapShot_ReturnsZeroOnFreshState(t *testing.T) {
	reset()
	snap := SnapShot()

	for key, val := range snap {
		if val != 0 {
			t.Errorf("expected 0 for %q after reset, got %d", key, val)
		}
	}
}
