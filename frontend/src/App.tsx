import {useEffect, useRef} from "react";


function App() {
  const socketRef  = useRef<WebSocket | null>(null);

  useEffect(() => {
    const socket = new WebSocket("ws://localhost:8080/ws");
    socketRef.current = socket;

    socket.onopen = () => {
      console.log("[WS] connected");
    };

    socket.onmessage = (event) => {
      console.log("[WS] message:", event.data);
    };

    socket.onclose = () => {
      console.log("[WS] disconnected");
    };

    socket.onerror = (err) => {
      console.error("[WS] error", err);
    };

    return () => {
      socket.close();
    };
  }, []);

  return (
    <div style={{padding: "24px", fontFamily: "sans-serif"}}>
      <h1>Semantic-Duel</h1>
      <p>Check console for WebSocket events</p>
    </div>
  );
}

export default App;
