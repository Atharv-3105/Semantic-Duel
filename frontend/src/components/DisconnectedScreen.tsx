interface Props {
    onReconnect: () => void;
}

export function DisconnectedScreen({onReconnect} : Props) {
    return (
        <div style={{marginTop: 40}}>
            <h2>Disconnected</h2>
            <p>Lost connection to server.</p>
            <button onClick = {onReconnect} style = {{padding: 8}}>
                Reconnect
            </button>
        </div>
    );
}