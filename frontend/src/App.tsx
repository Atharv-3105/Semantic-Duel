// import {useEffect, useRef} from "react";
import { useGameSocket } from "./hooks/useGameSocket";
import { WaitingScreen } from "./components/WaitingScreen";
import { GameHeader } from "./components/GameHeader";

function App() {
  const {phase, target, scores, winner, waitingMessage, secondsLeft} = useGameSocket();
  // const socketRef  = useRef<WebSocket | null>(null);

  return (
    <div style={{padding:"32", fontFamily:"sans-serif", maxWidth: 600}}>
      <h1>Semantic-Duel</h1>

      <p><strong>Phase:</strong>{phase}</p>

      {phase === "WAITING" && (
            <WaitingScreen message={waitingMessage} />
      )}

      {phase === "IN_GAME" && target && (
        <>
          <GameHeader target = {target} secondsLeft={secondsLeft} />
          <pre>{JSON.stringify(scores, null, 2)}</pre>
        </>
      )}

      {phase === "GAME_OVER" && (
        <>
          <h2>Game Over</h2>
          <p><strong>Winner:</strong>{winner || "Tie"}</p>
          <pre>{JSON.stringify(scores, null, 2)}</pre>
        </>
      )}


      {phase === "DISCONNECTED" && (
          <p>Disconnected from server.</p>
      )}
    </div>
  );
}

export default App;
