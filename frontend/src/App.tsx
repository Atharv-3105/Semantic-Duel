import { useGameSocket } from "./hooks/useGameSocket";
import { WaitingScreen } from "./components/WaitingScreen";
import { GameArena } from "./components/GameArena";
import { GameOverScreen } from "./components/GameOverScreen";
import { DisconnectedScreen } from "./components/DisconnectedScreen";
import { ConnectingScreen } from "./components/ConnectingScreen";
import "./App.css";

function App() {
  const {
    phase,
    target,
    scores,
    winner,
    waitingMessage,
    secondsLeft,
    duration,
    submitWord,
    reconnect,
    connected,
  } = useGameSocket();

  return (
    <div className="app-container">
      {/* Ambient background effects */}
      <div className="ambient-orb ambient-orb--cyan" />
      <div className="ambient-orb ambient-orb--purple" />

      {/* Header */}
      <header className="app-header">
        <h1 className="game-logo">
          <span className="logo-icon">⚔️</span>
          Semantic Duel
        </h1>
        <div className="header-meta">
          <span className={`status-badge ${connected ? "status-badge--connected" : "status-badge--disconnected"}`}>
            <span className="status-dot" />
            {connected ? "Connected" : "Disconnected"}
          </span>
        </div>
      </header>

      {/* Phase Content */}
      <main className="app-main">
        {phase === "CONNECTING" && <ConnectingScreen />}

        {phase === "WAITING" && (
          <WaitingScreen message={waitingMessage} />
        )}

        {phase === "IN_GAME" && target && (
          <GameArena
            target={target}
            secondsLeft={secondsLeft}
            duration={duration}
            scores={scores}
            disabled={!connected || phase !== "IN_GAME"}
            onSubmitWord={submitWord}
          />
        )}

        {phase === "GAME_OVER" && (
          <GameOverScreen winner={winner} scores={scores} />
        )}

        {phase === "DISCONNECTED" && (
          <DisconnectedScreen onReconnect={reconnect} />
        )}
      </main>

      {/* Footer */}
      <footer className="app-footer">
        <p>Submit words semantically related to the target — best score wins!</p>
      </footer>
    </div>
  );
}

export default App;
