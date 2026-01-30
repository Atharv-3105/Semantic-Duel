package ws

import "encoding/json"


//BASE ENVELOPE  for ALL INCOMING MESSAGES
type ClientMessage struct {
	Type    string `json:"type"`
	Payload json.RawMessage		`json:"payload"`
}

type ClientMessageType string 

const (
	ClientWordSubmit ClientMessageType = "WORD_SUBMIT"
)

type WordSubmitPayload struct {
	Word 	string  	`json:"word"`
}

//BASE ENVELOPE for ALLL OUTGOING MESSAGES
type EventMessage struct {
	Type   ServerEventType 		`json:"type"`
	Payload  any 				`json:"payload"`
}

type ServerEventType string

const (
	EventGameStart  ServerEventType = "GAME_START"
	EventScoreUpdate ServerEventType = "SCORE_UPDATE"
	EventGameOver	ServerEventType = "GAME_OVER"
)

type GameStartPayload struct {
	Target  string `json:"target"`
	Duration  int 	`json:"duration"`
}

type ScoreUpdatePayload struct {
	PlayerID  string 	`json:"playerId"`
	Score     int 		`json:"score"`
}

type GameOverPayload  struct {
	Winner	 string      `json:"winner"`
	Scores    map[string]int  `json:"scores"`
}

