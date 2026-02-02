interface Props {
    message?: string | null;
}

export function WaitingScreen({message}: Props) {
    return (
        <div style={{marginTop: 40}}>
            <h2>Waiting for opponent...</h2>
            <p>{message ?? "Hang tight. Matching you with another player."}</p>
        </div>
    );
}