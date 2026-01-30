package room

type EventType string 	

const (
	GameStart	EventType = "GAME_START"
	WordSubmit	EventType = "WORD_SUBMIT"
	ScoreUpdate	EventType = "SCORE_UPDATE"
	GameOver	EventType = "GAME_OVER"
)


type Event struct {
	Type    EventType 	`json:"type"`
	Payload	 interface{}  `json:"payload"`
}

