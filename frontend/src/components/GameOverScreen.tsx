interface Props {
  winner: string | null;
  scores: Record<string, number>;
}

export function GameOverScreen({ winner, scores }: Props) {
  const entries = Object.entries(scores);
  const isDraw = !winner;

  return (
    <div className="game-over-screen">
      <div className={`game-over-card glass-card ${isDraw ? "" : "game-over-card--win"}`}>
        {/* Trophy / Draw Icon */}
        <div className="trophy-icon">
          {isDraw ? "🤝" : "🏆"}
        </div>

        {/* Title */}
        <h2 className={`game-over-title ${isDraw ? "game-over-title--draw" : "game-over-title--win"}`}>
          {isDraw ? "It's a Draw!" : "Victory!"}
        </h2>

        <p className="game-over-subtitle">
          {isDraw
            ? "Both players fought valiantly — an even match!"
            : "A champion has emerged from the semantic arena."}
        </p>

        {/* Final Scores */}
        <div className="final-scores">
          <p className="final-scores-title">Final Scores</p>
          {entries.map(([playerId, score]) => {
            const isWinner = playerId === winner;
            return (
              <div
                key={playerId}
                className={`score-row ${isWinner ? "score-row--winner" : ""}`}
              >
                <div className="score-row__player">
                  <span>{playerId.slice(0, 8)}</span>
                  {isWinner && (
                    <span className="score-row__badge score-row__badge--winner">
                      Winner
                    </span>
                  )}
                </div>
                <span className="score-row__value">{score}</span>
              </div>
            );
          })}
        </div>

        {/* Next Match Indicator */}
        <div className="next-match">
          <div className="mini-spinner" />
          <span>Finding next match...</span>
        </div>
      </div>
    </div>
  );
}