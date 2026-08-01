# Semantic-Duel — Senior Developer Code Review

> **Reviewer perspective:** Senior Backend Engineer with experience in real-time systems, Go concurrency, and distributed architecture.
> **Date:** August 2026
> **Verdict:** A genuinely promising project with a clean conceptual foundation. It shows you understand the hardest parts of real-time multiplayer systems early. The gaps are mostly about production hardening, not fundamental design mistakes.

---

## 1. Technical Architecture

### System Overview

```
Browser (React/Vite)
    │  WebSocket (JSON)
    ▼
Go HTTP Server (:8080)
    ├── WS Hub         → manages raw connections (register/unregister)
    ├── Matchmaker     → FIFO queue, pairs players into Rooms
    ├── Room Manager   → tracks roomID → Room, clientID → Room
    └── Game Room      → owns game state, calls ML service, broadcasts events
            │  HTTP POST /similarity (2s timeout)
            ▼
Python FastAPI (:8001)
    └── spaCy en_core_web_md → cosine similarity
```

### Package Responsibility Map

| Package | Role | Verdict |
|---|---|---|
| `ws` | Transport layer — WebSocket lifecycle | ✅ Clean |
| `matchmaker` | Player pairing — FIFO queue | ✅ Simple and correct |
| `room` | Game state machine, scoring, I/O | ⚠️ Slightly overloaded |
| `game` | Pure scoring/state logic | ✅ Excellent isolation |
| `semantic` | HTTP client to Python service | ✅ Well-bounded |
| `config` | Env-var loading | ✅ Good |
| `metrics` | Atomic counters, JSON snapshot | ✅ Pragmatic |
| `logger` | Thin wrapper over `log.Logger` | ⚠️ Underpowered |
| `target` | Random word provider | ✅ Simple and testable |

### Data / Event Flow

```
Client sends WORD_SUBMIT
    → Hub.ReadPump receives raw bytes
    → ClientMessage unmarshalled in main.go goroutine
    → Room.HandleWord() called
        → rate limit check (in-memory map)
        → input validation (unicode letters, 0–32 chars)
        → HTTP POST to Python /similarity (blocks goroutine, 2s timeout)
        → SimilarityToScore() → best-score logic
        → SCORE_UPDATE broadcast to both clients
    → Timer goroutine fires (time.Sleep until EndsAt)
    → Room.endGame() → GAME_OVER broadcast
    → cleanupCh receives roomID
    → main.go re-queues both players after 2s sleep
```

---

## 2. Good Decisions ✅

### 2.1 Strict Layer Separation (Transport ≠ Game ≠ ML)
The `ws` package knows nothing about game rules. The `game` package knows nothing about WebSockets. The `semantic` package is a pure HTTP adapter. This is textbook separation of concerns and makes the system highly testable at each layer in isolation.

### 2.2 Server-Authoritative Design
All scoring happens on the server. The client **never** sends a score — only a word. This is the right call for anti-cheat. A naive implementation would have clients compute and report their own similarity scores; you avoided that trap.

### 2.3 Best-Score-Wins Semantics
```go
prev, exists := r.State.Scores[playerID]
if !exists || score > prev {
    r.State.Scores[playerID] = score
}
```
This is elegant. Players are rewarded for their best submission, not their last. It creates a good gameplay loop where early good guesses are never penalized.

### 2.4 Graceful Unregister with `recover()` in `Client.Send`
```go
defer func() {
    if r := recover(); r != nil { /* Ignore */ }
}()
```
The `send` channel can be closed by the Hub if the client disconnects. Without this guard, sending to a closed channel panics and crashes the goroutine. This is a classic WebSocket pitfall that you correctly handled.

### 2.5 Typed Message Protocol
Defining `ServerEventType`, `ClientMessageType`, and all payload structs as explicit Go types (`ws/messages.go`) means message formats are enforced by the compiler. Mirroring this with discriminated union types in TypeScript (`types/messages.ts`) makes the protocol self-documenting and eliminates a class of runtime bugs.

### 2.6 Atomic Metrics
Using `sync/atomic` for game counters is the correct, zero-allocation approach for hot-path counters. Using a mutex here would be overkill.

### 2.7 Configurable via Environment Variables
`config.go` provides clean defaults and reads from env vars. This is 12-factor app compliant and makes the service Docker/Kubernetes ready without code changes.

### 2.8 Ping/Pong Heartbeat Implementation
The `writeWait`, `pongWait`, `pingPeriod` constants and the `SetPongHandler` are implemented correctly. Many developers get this wrong (e.g., not resetting the read deadline on every pong). You got it right.

### 2.9 Python Service Design
Keeping ML inference in a separate Python process is the correct polyglot boundary. spaCy's `en_core_web_md` gives reasonable word vectors for a prototype without needing a GPU or a heavy transformer. Splitting `model.py`, `similarity.py`, and `main.py` shows awareness of single-responsibility even in small Python code.

### 2.10 `time.Until(r.State.EndsAt)` Pattern
```go
go func() {
    time.Sleep(time.Until(r.State.EndsAt))
    r.endGame()
}()
```
Smarter than just `time.Sleep(60 * time.Second)`. If a game is started with a delay for any reason, `time.Until` ensures the timer is still accurate to the wall-clock end time. Good precision thinking.

---

## 3. Bad Decisions / Bugs ❌

### 3.1 🔴 CRITICAL — Room Has No Mutex: Data Race on Game State
**File:** [`internal/room/room.go`](file:///C:/Temp/Semantic-Duel/internal/room/room.go)

`HandleWord` is called from the `main.go` goroutine (via the OnConnect channel goroutine). `endGame()` is called from a separate timer goroutine. Both read/write `r.State.Scores` and `r.State.Status` **concurrently without any synchronization**.

```go
// Goroutine 1 (message handler): writes r.State.Scores
r.State.Scores[playerID] = score

// Goroutine 2 (timer): reads r.State.Scores AND writes r.State.Status
for pid, score := range r.State.Scores { ... }
r.State.Status = game.Finished
```

This is an **undetected data race**. Under the Go memory model, this is undefined behavior. Run `go test -race ./...` and you will catch it. The fix is a `sync.RWMutex` on the `Room` struct.

---

### 3.2 🔴 CRITICAL — `Hub.Run()` Blocks on `OnConnect` / `OnDisconnect`

**File:** [`internal/ws/hub.go`](file:///C:/Temp/Semantic-Duel/internal/ws/hub.go)

```go
case client := <-h.register:
    h.clients[client] = true
    h.OnConnect <- client   // ← BLOCKS until consumer reads
```

`OnConnect` is an **unbuffered channel**. If the goroutine in `main.go` that reads from `OnConnect` is busy (e.g., doing anything slow), `Hub.Run()` deadlocks and **no other client can connect or disconnect** until the channel clears. The same applies to `OnDisconnect`. This is a correctness bug that will manifest under any real load.

**Fix:** Either buffer the channels `make(chan *Client, 64)` or dispatch in a `go func()`.

---

### 3.3 🔴 CRITICAL — `HandleWord` Blocks the Message Dispatch Goroutine on HTTP

**File:** [`internal/room/room.go`](file:///C:/Temp/Semantic-Duel/internal/room/room.go), line 107

```go
sim, err := r.semantic.Similarity(word, r.State.TargetWord)
```

The HTTP call to the Python service (up to 2 seconds) happens **synchronously inside the message handler goroutine**. Because messages from all clients in a room funnel through the same goroutine (see the `OnConnect` handler in `main.go`), one slow HTTP call stalls all word submissions from that room. Under high load, if the Python service is slow, the entire match effectively pauses. This call must be dispatched in its own goroutine.

---

### 3.4 🟠 Module Name Mismatch
**File:** [`go.mod`](file:///C:/Temp/Semantic-Duel/go.mod)

```
module github.com/Atharv-3105/Graph-Duel
```

The module is named `Graph-Duel` but the repository and the project are called `Semantic-Duel`. This is a leftover from a rename/fork and will confuse any contributor who clones the repo. All imports internally use the wrong name.

---

### 3.5 🟠 `CleanupRoom` Called Twice on Disconnect
**File:** [`internal/room/manager.go`](file:///C:/Temp/Semantic-Duel/internal/room/manager.go), line 65

```go
func (m *Manager) HandleDisconnect(clientID string) {
    ...
    room.ForceEnd(clientID)   // calls r.cleanup() → sends roomID into cleanupCh
    m.CleanupRoom(room.ID)    // also deletes from the maps immediately
}
```

`ForceEnd` triggers `r.cleanup()`, which sends the room ID to `cleanupCh`. The `main.go` goroutine listening on `cleanupCh` then calls `roomManager.CleanupRoom(roomID)` **again**. Meanwhile, `HandleDisconnect` already called `CleanupRoom` directly. The second call on a disconnect is harmless only because `CleanupRoom` checks `if !ok { return nil, nil }`, but it then attempts to re-queue `nil` players. The `main.go` guards against `nil` but the path is fragile and confusing.

---

### 3.6 🟠 `rateLimitSeconds` is Passed but Never Used
**File:** [`internal/room/room.go`](file:///C:/Temp/Semantic-Duel/internal/room/room.go)

```go
type Room struct {
    ...
    rateLimitSeconds int  // ← stored
}
```

```go
// But the check hardcodes 1 second:
if ok && now.Sub(last) < time.Second {
```

The configurable `rateLimitSeconds` field is populated from config and passed to `NewRoom`, but the actual rate limit check ignores it and uses a hardcoded `time.Second`. This means changing `RATE_LIMIT_SECONDS` in config has no effect.

---

### 3.7 🟠 `game.StartGame` Hardcodes 60 Seconds
**File:** [`internal/game/engine.go`](file:///C:/Temp/Semantic-Duel/internal/game/engine.go)

```go
func StartGame(state *State) {
    state.Status = Active
    state.EndsAt = time.Now().Add(60 * time.Second) // ← hardcoded
}
```

`gameDuration` is correctly stored in `Room` and sent to the client in `GAME_START`, but `StartGame` ignores it and always sets a 60-second game. This means changing `GAME_DURATION_SECONDS` only affects the countdown shown to the client — the server-side timer still runs for 60 seconds. A subtle and misleading bug.

---

### 3.8 🟡 `WAITING` Server Event Never Sent
**File:** [`frontend/src/hooks/useGameSocket.ts`](file:///C:/Temp/Semantic-Duel/frontend/src/hooks/useGameSocket.ts), line 44

The frontend handles a `"WAITING"` server message, but the Go backend **never sends this message type**. The frontend assumes `WAITING` phase based on `socket.onopen`, but there's no server-push for "still waiting for opponent." If a player waits a long time, they have no server-confirmed feedback. The `WaitingMessage` interface in TypeScript types is dead code.

---

### 3.9 🟡 `handleMessage` Dead Code in ws/handler.go
**File:** [`internal/ws/handler.go`](file:///C:/Temp/Semantic-Duel/internal/ws/handler.go), line 34

```go
func (c *Client) handleMessage(m IncomingMessage) {
    log.Println("[WS] message received:", m.Type, m.Word)
}
```

This method is never called. The `IncomingMessage` struct is also unused — the actual message parsing uses `ClientMessage`. This is dead code from an earlier iteration that was never cleaned up.

---

### 3.10 🟡 `active_connections` in Metrics Snapshot is Missing
**File:** [`internal/metrics/metrics.go`](file:///C:/Temp/Semantic-Duel/internal/metrics/metrics.go)

The README documents `"active_connections"` as a metrics field, but the `SnapShot()` function only returns 3 keys: `games_started`, `games_completed`, `disconnects`. The documented field is missing from the implementation.

---

### 3.11 🟡 Frontend WebSocket URL is Hardcoded
**File:** [`frontend/src/hooks/useGameSocket.ts`](file:///C:/Temp/Semantic-Duel/frontend/src/hooks/useGameSocket.ts), line 19

```ts
const socket = new WebSocket("ws://localhost:8080/ws");
```

Hardcoded to `localhost`. This frontend cannot be deployed to any environment without a code change. It should read from `import.meta.env.VITE_WS_URL`.

---

### 3.12 🟡 `hub.go` — Hub Clients Map is Not Goroutine-Safe

**File:** [`internal/ws/hub.go`](file:///C:/Temp/Semantic-Duel/internal/ws/hub.go)

The `clients` map is only ever accessed from `Hub.Run()` which runs in a single goroutine — so this is actually fine as written. But `len(h.clients)` is logged as a metric proxy. If you ever expose `clients` directly (e.g., for broadcasting), this will become a race. Add a note or keep access strictly within `Run`.

---

## 4. Improvements to Make

### 4.1 Add a `sync.RWMutex` to Room (Fix the Data Race)

```go
type Room struct {
    mu   sync.RWMutex
    // ... other fields
}

func (r *Room) HandleWord(playerID, word string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    // ... now safe
}

func (r *Room) endGame() {
    r.mu.Lock()
    defer r.mu.Unlock()
    // ... now safe
}
```

### 4.2 Dispatch ML Call in a Goroutine

```go
func (r *Room) HandleWord(playerID, word string) {
    // ... validate synchronously
    
    go func() {
        sim, err := r.semantic.Similarity(word, r.State.TargetWord)
        if err != nil { return }
        
        r.mu.Lock()
        defer r.mu.Unlock()
        // ... update score
    }()
}
```

### 4.3 Use `gameDuration` in `StartGame`

```go
// In game/engine.go
func StartGame(state *State, durationSeconds int) {
    state.Status = Active
    state.EndsAt = time.Now().Add(time.Duration(durationSeconds) * time.Second)
}
```

### 4.4 Use `rateLimitSeconds` in `HandleWord`

```go
if ok && now.Sub(last) < time.Duration(r.rateLimitSeconds)*time.Second {
    return
}
```

### 4.5 Fix the Module Name

```bash
# go.mod
module github.com/Atharv-3105/Semantic-Duel
```

Then update all imports. A simple find-and-replace across the project.

### 4.6 Add `active_connections` Metric

```go
// metrics.go
var ActiveConnections int64

func IncConnections()  { atomic.AddInt64(&ActiveConnections, 1) }
func DecConnections()  { atomic.AddInt64(&ActiveConnections, -1) }

func SnapShot() map[string]int64 {
    return map[string]int64{
        "games_started":      atomic.LoadInt64(&GamesStarted),
        "games_completed":    atomic.LoadInt64(&GamesCompleted),
        "disconnects":        atomic.LoadInt64(&Disconnects),
        "active_connections": atomic.LoadInt64(&ActiveConnections),
    }
}
```

### 4.7 Externalize the WebSocket URL in Frontend

```ts
// .env
VITE_WS_URL=ws://localhost:8080/ws

// useGameSocket.ts
const wsUrl = import.meta.env.VITE_WS_URL ?? "ws://localhost:8080/ws";
const socket = new WebSocket(wsUrl);
```

### 4.8 Add a `WAITING_FOR_OPPONENT` Server Push

Send a message when a player is first enqueued so the client gets server confirmation:
```go
// In matchmaker, after enqueue but before pairing:
client.Send(ws.EventWaiting, ws.WaitingPayload{Message: "Waiting for opponent..."})
```

### 4.9 Upgrade the Logger

The current logger is a thin wrapper around `log.Logger` with no log levels, no structured output, and no way to suppress debug noise in production. Consider adopting `log/slog` (stdlib in Go 1.21+) which gives you structured JSON logging for free:

```go
import "log/slog"
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
logger.Info("client connected", "client_id", id, "active", count)
```

### 4.10 Model Caching in Python Service

Currently, both words are re-embedded on every request:
```python
v1 = embed(req.word)    # nlp(req.word).vector
v2 = embed(req.target)  # nlp(req.target).vector
```
The target word is the **same for the entire game**. Cache the target embedding in a `dict` keyed by word string. This halves inference time per request for the common case.

```python
from functools import lru_cache

@lru_cache(maxsize=512)
def embed_cached(text: str):
    return nlp(text).vector
```

### 4.11 Add `context.Context` to `semantic.Client.Similarity`

The 2s timeout is on the `http.Client`. But there's no way to cancel an in-flight request if the game ends mid-inference. Add context propagation:

```go
func (c *Client) Similarity(ctx context.Context, word, target string) (float64, error) {
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, ...)
    resp, err := c.client.Do(req)
    ...
}
```

### 4.12 Write More Tests

Current coverage is minimal:
- `scoring_test.go` — tests `SimilarityToScore` (good, but boundary cases only)
- `room_test.go` — tests `UpdateScore` best-score logic (good start)

**Missing test coverage:**
- `HandleWord` rate limiting logic
- `HandleWord` input validation (empty, too long, non-letters)
- `matchmaker.Enqueue` — pairing and room creation
- `endGame` winner determination (tie-breaking, no submissions case)
- `target.Provider.Random` with empty word list

### 4.13 Add a Dockerfile and docker-compose

The project has two services (Go + Python) that need to be coordinated. A `docker-compose.yml` would make the developer experience dramatically better:

```yaml
services:
  game-server:
    build: .
    ports: ["8080:8080"]
    environment:
      SEMANTIC_SERVICE_URL: http://semantic:8001
    depends_on: [semantic]

  semantic:
    build: ./semantic-service
    ports: ["8001:8001"]
```

---

## 5. Summary Scorecard

| Category | Score | Notes |
|---|---|---|
| **Architecture / SoC** | 9/10 | Excellent layer separation |
| **Correctness / Safety** | 5/10 | Data race, blocking HTTP in handler, double-cleanup |
| **Go Idioms** | 7/10 | Good channel use; missing mutex; mixed logger |
| **Configurability** | 6/10 | Good env config but 2 values are hardcoded or ignored |
| **Testing** | 4/10 | Very sparse; critical paths untested |
| **Frontend** | 6/10 | Typed protocol is great; hardcoded URL, no styling |
| **Documentation** | 8/10 | README is well-written; architecture diagram is accurate |
| **Production Readiness** | 4/10 | No Docker, no graceful shutdown, no circuit breaker |

---

## 6. Final Take

You've built something that's architecturally more mature than most first-pass real-time games. The choice to use Go for the game server, isolate ML in Python, and implement a clean message-envelope protocol shows good systems thinking.

The critical issues (data race, blocking HTTP in the dispatch goroutine, config values being ignored) are the kind of bugs that are invisible in local dev and catastrophic under production load. Fix those first.

The low test coverage is the other major gap. The `room` package especially — which has the most complex logic — has almost no tests. If you add a mutex and restructure `HandleWord` to async, tests will be what saves you from regressions.

This is a strong foundation. With 2–3 days of focused hardening work it would be a genuinely solid real-time server.
