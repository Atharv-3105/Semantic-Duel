import {useEffect, useRef} from "react";
import { useGameSocket } from "./hooks/useGameSocket";

function App() {
  const {phase, target, scores, winner, waitingMessage} = useGameSocket();
  // const socketRef  = useRef<WebSocket | null>(null);

  return (
    <div style={{padding:"24px", fontFamily:"sans-serif"}}>
      <h1>Semantic-Duel</h1>

      <p><strong>Phase:</strong>{phase}</p>

      {phase === "WAITING" && <p>{waitingMessage}</p>}

      {phase === "IN_GAME" && (
        <>
          <p><strong>Target:</strong>{target}</p>
          <pre>{JSON.stringify(scores, null, 2)}</pre>
        </>
      )}

      {phase === "GAME_OVER" && (
        <>
          <p><strong>Winner:</strong>{winner || "Tie"}</p>
          <pre>{JSON.stringify(scores, null, 2)}</pre>
        </>
      )}


      {phase === "DISCONNECTED" && <p>Disconnected</p>}
    </div>
  );
}

export default App;
