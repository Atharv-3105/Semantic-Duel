import { useEffect, useRef, useState, useCallback } from "react";
import type { GamePhase } from "../types/game";
import type { ServerMessage } from "../types/messages";

const WS_URL = import.meta.env.VITE_WS_URL ?? "ws://localhost:8080/ws";

export function useGameSocket() {
  const socketRef = useRef<WebSocket | null>(null);

  const [phase, setPhase] = useState<GamePhase>("CONNECTING");
  const [target, setTarget] = useState<string | null>(null);
  const [scores, setScores] = useState<Record<string, number>>({});
  const [winner, setWinner] = useState<string | null>(null);
  const [waitingMessage, setWaitingMessage] = useState<string | null>(null);
  const [secondsLeft, setSecondsLeft] = useState<number>(0);
  const [duration, setDuration] = useState<number>(60);
  const [connected, setConnected] = useState(false);

  const handleServerMessage = useCallback((msg: ServerMessage) => {
    switch (msg.type) {
      case "WAITING":
        setWaitingMessage(msg.payload.message);
        setPhase("WAITING");
        break;

      case "GAME_START":
        setTarget(msg.payload.target);
        setScores({});
        setWinner(null);
        setDuration(msg.payload.duration);
        setSecondsLeft(msg.payload.duration);
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
  }, []);

  const connect = useCallback(() => {
    const socket = new WebSocket(WS_URL);
    socketRef.current = socket;

    socket.onopen = () => {
      console.log("[WS] connected");
      setConnected(true);
      setPhase("WAITING");
    };

    socket.onmessage = (event) => {
      const msg: ServerMessage = JSON.parse(event.data);
      handleServerMessage(msg);
    };

    socket.onclose = () => {
      console.log("[WS] disconnected");
      setConnected(false);
      setPhase("DISCONNECTED");
    };

    socket.onerror = () => {
      console.log("[WS] error");
    };

    return socket;
  }, [handleServerMessage]);

  useEffect(() => {
    const socket = connect();
    return () => socket.close();
  }, [connect]);

  // Countdown timer
  useEffect(() => {
    if (phase !== "IN_GAME") return;

    const timer = setInterval(() => {
      setSecondsLeft((prev) => (prev > 0 ? prev - 1 : 0));
    }, 1000);

    return () => clearInterval(timer);
  }, [phase]);

  function submitWord(word: string) {
    if (!socketRef.current) return;
    if (phase !== "IN_GAME") return;

    socketRef.current.send(
      JSON.stringify({
        type: "WORD_SUBMIT",
        payload: { word },
      })
    );
  }

  function reconnect() {
    if (socketRef.current) {
      socketRef.current.close();
    }
    setPhase("CONNECTING");
    connect();
  }

  return {
    phase,
    target,
    scores,
    winner,
    waitingMessage,
    secondsLeft,
    duration,
    submitWord,
    connected,
    reconnect,
  };
}