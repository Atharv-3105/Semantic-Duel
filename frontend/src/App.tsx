// import {useEffect, useRef} from "react";
import { useGameSocket } from "./hooks/useGameSocket";
import { WaitingScreen } from "./components/WaitingScreen";
import { GameHeader } from "./components/GameHeader";
import { WordInput } from "./components/WordInput";
import { Scoreboard } from "./components/Scoreboard";
import { GameOverScreen } from "./components/GameOverScreen";
import { DisconnectedScreen } from "./components/DisconnectedScreen";

function App() {
  const {phase, target, scores, winner, waitingMessage, secondsLeft, submitWord, reconnect, connected} = useGameSocket();
  // const socketRef  = useRef<WebSocket | null>(null);

  return (
    <div style={{padding:"32", fontFamily:"sans-serif", maxWidth: 600}}>
      <h1>Semantic-Duel</h1>

      <p><strong>Phase:</strong>{phase}</p>
      <p style={{opacity: 0.6}}>Status: {connected ? "Connected" : "Disconnected"}</p>

      {phase === "WAITING" && (
            <WaitingScreen message={waitingMessage} />
      )}

      {phase === "IN_GAME" && target && (
        <>
          <GameHeader target = {target} secondsLeft={secondsLeft} />
          <WordInput disabled = {!connected || phase !== "IN_GAME"} onSubmit={submitWord}/>
          <Scoreboard scores = {scores} />
        </>
      )}

      {phase === "GAME_OVER" && (
        <GameOverScreen winner = {winner} scores = {scores} />
      )}


      {phase === "DISCONNECTED" && (
          <DisconnectedScreen onReconnect={reconnect} />
      )}
    </div>
  );
}

export default App;
