interface Props{
    target: string;
    secondsLeft: number;
}

export function GameHeader({target, secondsLeft} : Props) {
    return (
        <div style={{marginBottom: 24}}>
            <h2>Target Word: <span style={{color: "#2563eb"}}>{target}</span></h2>
            <p>Time Left: {secondsLeft}</p>
        </div>
    );
}