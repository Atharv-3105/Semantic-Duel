export function ConnectingScreen() {
  return (
    <div className="connecting-screen animate-fade-in">
      <div className="connecting-card glass-card">
        <div className="connecting-spinner">
          <div className="spinner-ring" />
        </div>
        <h2 className="connecting-title">Establishing Connection</h2>
        <p className="connecting-subtitle">Connecting to the arena server...</p>
      </div>
    </div>
  );
}
