interface Props {
    scores: Record<string, number>;
}

export function Scoreboard({scores} : Props) {
    const entries = Object.entries(scores);

    if (entries.length === 0) {
        return <p>No scores yet</p>
    }

    return (
        <div style={{marginTop: 16}}>
            <h3>Scores</h3>
            <ul>
                {entries.map(([playerId, score]) => (
                    <li key = {playerId}>
                        {playerId.slice(0,6)}... : {score}
                    </li>
                ))}
            </ul>
        </div>
    );
}