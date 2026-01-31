export type ServerMessage = 
  | WaitingMessage
  | GameStartMessage
  | ScoreUpdateMessage
  | GameOverMessage;

export interface WaitingMessage {
    type: "WAITING";
    payload: {
        message: string;
    };
}

export interface GameStartMessage {
    type: "GAME_START";
    payload: {
        target: string;
        duration: number;
    };
}


export interface ScoreUpdateMessage {
    type: "SCORE_UPDATE";
    payload: {
        playerId: string;
        score: number;
    };
}


export interface GameOverMessage {
    type: "GAME_OVER";
    payload: {
        winner: string;
        scores: Record<string, number>;
    };
}