import { useState, useRef, useEffect } from "react";

interface Props {
  target: string;
  secondsLeft: number;
  duration: number;
  scores: Record<string, number>;
  disabled: boolean;
  onSubmitWord: (word: string) => void;
}

export function GameArena({
  target,
  secondsLeft,
  duration,
  scores,
  disabled,
  onSubmitWord,
}: Props) {
  const [value, setValue] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);
  const entries = Object.entries(scores);
  const [animatingPlayer, setAnimatingPlayer] = useState<string | null>(null);
  const prevScoresRef = useRef<Record<string, number>>({});

  // Detect score changes and trigger animation
  useEffect(() => {
    const prev = prevScoresRef.current;
    for (const [pid, score] of Object.entries(scores)) {
      if (prev[pid] !== undefined && prev[pid] !== score) {
        setAnimatingPlayer(pid);
        const timer = setTimeout(() => setAnimatingPlayer(null), 400);
        return () => clearTimeout(timer);
      }
    }
    prevScoresRef.current = { ...scores };
  }, [scores]);

  // Auto-focus input when game starts
  useEffect(() => {
    if (!disabled && inputRef.current) {
      inputRef.current.focus();
    }
  }, [disabled]);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const word = value.trim().toLowerCase();
    if (!word) return;
    onSubmitWord(word);
    setValue("");
  }

  // Timer calculations
  const circumference = 2 * Math.PI * 26;
  const progress = duration > 0 ? secondsLeft / duration : 0;
  const dashOffset = circumference * (1 - progress);
  const isDanger = secondsLeft <= 10;

  // Player entries: first is "you", second is "opponent"
  const player1 = entries[0];
  const player2 = entries[1];

  return (
    <div className="game-arena">
      {/* ─── Target Word ─── */}
      <div className="target-section glass-card">
        <p className="target-label">Target Word</p>
        <h2 className="target-word">{target}</h2>

        {/* Timer Ring */}
        <div className={`timer-container ${isDanger ? "timer-container--danger" : ""}`}>
          <svg className="timer-svg" viewBox="0 0 56 56">
            <circle className="timer-bg" cx="28" cy="28" r="26" />
            <circle
              className="timer-progress"
              cx="28"
              cy="28"
              r="26"
              strokeDasharray={circumference}
              strokeDashoffset={dashOffset}
            />
          </svg>
          <span className="timer-value">{secondsLeft}s</span>
        </div>
      </div>

      {/* ─── Player Scores ─── */}
      <div className="players-section">
        <div className="player-card glass-card player-card--you">
          <span className="player-label">You</span>
          <span className="player-id">
            {player1 ? player1[0].slice(0, 8) : "—"}
          </span>
          <span
            className={`player-score ${animatingPlayer === player1?.[0] ? "player-score--animate" : ""}`}
          >
            {player1 ? player1[1] : 0}
          </span>
          <div className="score-bar">
            <div
              className="score-bar__fill"
              style={{ width: `${player1 ? player1[1] : 0}%` }}
            />
          </div>
        </div>

        <div className="vs-divider">
          <span className="vs-text">VS</span>
        </div>

        <div className="player-card glass-card player-card--opponent">
          <span className="player-label">Opponent</span>
          <span className="player-id">
            {player2 ? player2[0].slice(0, 8) : "—"}
          </span>
          <span
            className={`player-score ${animatingPlayer === player2?.[0] ? "player-score--animate" : ""}`}
          >
            {player2 ? player2[1] : 0}
          </span>
          <div className="score-bar">
            <div
              className="score-bar__fill"
              style={{ width: `${player2 ? player2[1] : 0}%` }}
            />
          </div>
        </div>
      </div>

      {/* ─── Word Input ─── */}
      <div className="input-section">
        <form className="input-form" onSubmit={handleSubmit}>
          <input
            ref={inputRef}
            id="word-input"
            type="text"
            className="input-field"
            value={value}
            disabled={disabled}
            onChange={(e) => setValue(e.target.value)}
            placeholder="Type a word..."
            autoComplete="off"
            spellCheck={false}
            maxLength={32}
          />
          <button
            type="submit"
            className="submit-btn"
            disabled={disabled || !value.trim()}
            id="submit-word-button"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <line x1="22" y1="2" x2="11" y2="13" />
              <polygon points="22 2 15 22 11 13 2 9 22 2" />
            </svg>
            Submit
          </button>
        </form>
        <p className="input-hint">
          Type a word semantically related to "{target}" — best score wins!
        </p>
      </div>
    </div>
  );
}
