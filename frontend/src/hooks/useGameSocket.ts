import { useEffect, useRef, useState } from "react";
import type { GamePhase } from "../types/game";
import type { ServerMessage } from "../types/messages";


export function useGameSocket() {
    const socketRef = useRef<WebSocket | null>(null);

    const [phase, setPhase] = useState<GamePhase>("CONNECTING");
    const [target, setTarget] = useState<string | null>(null);
    const [scores, setScores] = useState<Record<string, number>>({});
    const [winner, setWinner] = useState<string | null>(null);
    const [waitingMessage, setWaitingMessage] = useState<string | null>(null);

    useEffect(() => {
        const socket = new WebSocket("ws://localhost:8080/ws");
        socketRef.current = socket;

        socket.onopen = () => {
            console.log("[WS] connected");
            setPhase("WAITING");
        };

        socket.onmessage = (event) => {
            const msg: ServerMessage = JSON.parse(event.data);
            handleServerMessage(msg);
        };

        socket.onclose = () => {
            console.log("[WS] disconnected");
            setPhase("DISCONNECTED");
        };

        return () => socket.close();
    }, []);

    function handleServerMessage(msg : ServerMessage) {
        switch(msg.type) {
            case "WAITING":
                setWaitingMessage(msg.payload.message);
                setPhase("WAITING");
                break;
            
            case "GAME_START":
                setTarget(msg.payload.target);
                setScores({});
                setWinner(null);
                setPhase("IN_GAME");
                break;

            case "SCORE_UPDATE":
                setScores((prev) => ({
                    ...prev,
                    [msg.payload.playerId]: msg.payload.score,
                }));
                break;
            
            case "GAME_OVER":
                setScores(msg.payload.scores);
                setWinner(msg.payload.winner);
                setPhase("GAME_OVER");
                break;
        }
    }

    return {
        phase,
        target,
        scores,
        winner,
        waitingMessage,
        scoket: socketRef,
    };
}