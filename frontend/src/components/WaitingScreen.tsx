interface Props {
  message?: string | null;
}

export function WaitingScreen({ message }: Props) {
  return (
    <div className="waiting-screen animate-fade-in">
      <div className="waiting-card glass-card">
        <div className="waiting-radar">
          <div className="radar-ring radar-ring--1" />
          <div className="radar-ring radar-ring--2" />
          <div className="radar-ring radar-ring--3" />
          <div className="radar-dot" />
        </div>
        <h2 className="waiting-title">Searching for Opponent</h2>
        <p className="waiting-subtitle">
          {message ?? "Scanning the arena for a worthy challenger..."}
        </p>
        <div className="waiting-dots">
          <span className="dot dot--1" />
          <span className="dot dot--2" />
          <span className="dot dot--3" />
        </div>
      </div>
    </div>
  );
}