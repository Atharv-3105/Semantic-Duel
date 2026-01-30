# Graph-Duel 
**Real-time Multiplayer Semantic Word-Battle Backend**

Graph-Duel is a production-grade, real-time multiplayer backend where two players compete by submitting words that are semantically related to a target word.  
The server scores submissions using a lightweight ML service and enforces all game rules authoritatively.
---

## ✨ Key Features

- Real-time multiplayer gameplay using WebSockets
- Server-authoritative game logic
- Lightweight semantic similarity scoring (CPU-only)
- Strict message protocol and state gating
- Abuse prevention (rate limiting, validation)
- Graceful disconnect handling
- Automatic room cleanup and requeue
- Config-driven runtime (no hardcoded values)
- Observability via metrics and structured logs
- Graceful shutdown on SIGINT / SIGTERM

---

## 🧠 High-Level Architecture

Clients (WebSocket)

|

v

+----------------------+

| WebSocket Hub |

| (Connection Mgmt) |

+----------------------+


|

v

+----------------------+

| Matchmaker |

| Room Manager |

+----------------------+


|

v

+----------------------+

| Game Room |

| (Authoritative State)|

+----------------------+


|

v

+----------------------+

| Semantic ML Service |

| (Python, CPU-only) |
+----------------------+


---

## Design Principles

- Strict separation of concerns (transport ≠ game ≠ ML)
- Event-driven lifecycle
- Server-authoritative state
- Best-effort messaging with strong state validation
- Explicit cleanup and ownership
- No hidden globals or side effects

---

## 🔌 Message Protocol (Simplified)

### Client → Server

```json
{
  "type": "WORD_SUBMIT",
  "payload": {
    "word": "fire"
  }
}
