package metrics

import  "sync/atomic"

var GamesStarted int64
var GamesCompleted int64
var Disconnects int64

func IncGamesStarted() {
	atomic.AddInt64(&GamesStarted, 1)
}

func IncGamesCompleted() {
	atomic.AddInt64(&GamesCompleted, 1)
}

func IncDisconnects() {
	atomic.AddInt64(&Disconnects, 1)
}

func SnapShot() map[string]int64 {
	return map[string]int64 {
		"games_started": atomic.LoadInt64(&GamesStarted),
		"games_completed": atomic.LoadInt64(&GamesCompleted),
		"disconnects":		atomic.LoadInt64(&Disconnects),
	}
}