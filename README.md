# Semantic-Duel 🧠⚔️

[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![WebSockets](https://img.shields.io/badge/WebSockets-Real--Time-FF6B6B?style=flat)](https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API)

> **Real-time multiplayer semantic word battle game** — Compete head-to-head by submitting words semantically related to a target word. Powered by Go, WebSockets, and lightweight ML inference.

---

## 🎮 How It Works

Two players enter a **60-second duel**. A target word (e.g., *"ocean"*) is revealed. Both players race to submit semantically related words (*"wave"*, *"tide"*, *"blue"*). A lightweight ML service scores each submission by semantic similarity. **Highest score wins.**


---

## ✨ Key Features

| Feature | Implementation |
|---------|---------------|
| **Real-time Gameplay** | Gorilla WebSockets with ping/pong heartbeats |
| **Server-Authoritative** | All scoring and state managed server-side (anti-cheat) |
| **Concurrent Architecture** | Lock-free message passing via Go channels |
| **Rate Limiting** | 1-second cooldown per player submission |
| **Graceful Degradation** | Circuit-breaker ready ML service client |
| **Automatic Requeue** | Players auto-matched to new games after completion |
| **Input Validation** | Unicode normalization, length limits, character filtering |

---

## 🏗️ Architecture

```
┌─────────────┐      WebSocket       ┌─────────────────┐
│   Client    │◄────────────────────►│   WS Hub        │
│  (Browser)  │   JSON Protocol      │  (Connection    │
└─────────────┘                      │   Management)   │
└────────┬────────┘
│
┌────────▼────────┐
│   Matchmaker    │
│  (FIFO Queue)   │
└────────┬────────┘
│
┌────────▼────────┐
│   Game Room     │
│  (Authoritative │
│   State Machine)│
└────────┬────────┘
│
┌────────▼────────┐
│  ML Service     │
│  (Python/FastAPI│
│   CPU-only)     │
└─────────────────┘
```
---

### Design Principles

- **Transport ≠ Game ≠ ML**: Strict separation of concerns
- **Event-Driven Lifecycle**: Channel-based communication over shared memory
- **Explicit Ownership**: No hidden globals; dependency injection throughout
- **Fail Fast**: 2-second timeout on ML inference; graceful fallbacks

---
---
## 🔧 Project Structure
```
Semantic-Duel/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/          # Environment configuration
│   ├── game/            # Scoring logic & game rules
│   ├── logger/          # Structured logging
│   ├── matchmaker/      # Player queue & pairing
│   ├── metrics/         # Prometheus-style counters
│   ├── room/            # Game room lifecycle & state
│   ├── semantic/        # ML service HTTP client
│   ├── target/          # Target word provider
│   └── ws/              # WebSocket hub & client management
├── semantic-service/    # Python FastAPI ML inference
│   ├── app/
│   └── requirements.txt
└── frontend/            # Static HTML/JS client
```
---
## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Python 3.9+ (for ML service)
- Redis (optional, for production scaling)

### 1. Clone & Setup

```bash
git clone https://github.com/Atharv-3105/Semantic-Duel.git
cd Semantic-Duel
```

### 2. Start the ML Service
```bash
cd semantic-service
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate
pip install -r requirements.txt
uvicorn app.main:app --host 0.0.0.0 --port 8001
```
### 3. Start the Game Server
```bash
# From project root
go run cmd/server/main.go
Server starts on :8080 by default.
```
### 4. Test Connection
```bash
# Connect via WebSocket
wscat -c ws://localhost:8080/ws
```

---
## 📊 Metrics Endpoint
```bash
curl http://localhost:8080/metrics
```
Returns:
```JSON
{
  "games_started": 42,
  "games_completed": 38,
  "disconnects": 4,
  "active_connections": 12
}
```
---
## 🛡️ Anti-Cheat Measures
| Attack Vector | 	Defense |
|---------|---------------|
| **Speed Hacking**	| Server-enforced 1s rate limit per player|
| **Score Spoofing** | 	All similarity scores computed server-side|
| **Replay Attacks**	| Game state machine rejects submissions after GAME_OVER|
| **Invalid Input**	| Unicode letter filtering, length limits (0-32 chars)|
| **Connection Fraud** |	TCP keepalive + application-level ping/pong (60s)|
