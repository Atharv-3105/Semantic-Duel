interface Props {
    winner: string | null;
    scores: Record<string, number>;
}


export function GameOverScreen({winner, scores} : Props) {
    const entries = Object.entries(scores);

    return (
        <div style = {{marginTop: 32}}>
            <h2>Game Over</h2>

            <p>
                <strong>Result:</strong>{" "}
                {winner ? `${winner.slice(0,6)}... wins` : "Draw"}
            </p>

            <h3>Final Scores</h3>
            <ul>
                {entries.map(([playerId, score]) => (
                    <li key = {playerId}>
                        {playerId.slice(0,6)}... : {score}
                    </li>
                ))}
            </ul>

            <p style = {{marginTop: 16, opacity: 0.7}}>
                Re-queueing for next match...
            </p>
        </div>
    );
}