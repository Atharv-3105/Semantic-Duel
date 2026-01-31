import { useEffect, useRef, useState } from "react";
import { GamePhase } from "../types/game";
import { ServerMessage } from "../types/messages";


export function useGameSocket() {
    const socketRef = useRef<WebSocket | null>(null);

    const [phase, setPhase] = useState<GamePhase>("CONNECTING");
    const [target, setTarget] = useState<string | null>(null);
    const [scores, setScores] = useState<Record<string, number>>({});
    const [winner, setWinner] = useState<string | null>(null);
    const [waitingMessage, setWaitingMessage] = useState<string | null>(null);

    useEffect(() => {
        const socket = new WebSocket()
    })
}