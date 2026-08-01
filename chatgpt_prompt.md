# ChatGPT Prompt — Semantic-Duel Bug Fixes + Pre-computed Embeddings

---

## CONTEXT

I have a real-time multiplayer word game called **Semantic-Duel**. Two players compete to submit words semantically related to a target word. The server scores submissions using ML similarity. It is written in Go (backend) + React/TypeScript (frontend) + Python FastAPI (ML service).

I need you to:
1. **Fix all known bugs** in the Go backend (listed below with current code)
2. **Replace the Python ML microservice** with a pre-computed embeddings approach — a one-time Python script that generates embeddings, and a Go `LocalClient` that loads them at startup and does similarity lookups with zero network calls.

After your changes, the only external dependency will be the Go binary + an embeddings file. No Python service needs to run at runtime.

---

## PROJECT STRUCTURE

```
Semantic-Duel/
├── cmd/server/main.go
├── go.mod                          ← module is wrongly named Graph-Duel, must be fixed
├── internal/
│   ├── config/config.go
│   ├── game/engine.go
│   ├── game/scoring.go
│   ├── logger/logger.go
│   ├── matchmaker/matchmaker.go
│   ├── metrics/metrics.go
│   ├── room/room.go
│   ├── room/manager.go
│   ├── room/state.go
│   ├── room/events.go
│   ├── semantic/client.go          ← replace this with LocalClient
│   ├── target/provider.go
│   ├── target/words.go
│   └── ws/
│       ├── hub.go
│       ├── client.go
│       ├── handler.go
│       └── messages.go
└── scripts/
    └── precompute.py               ← CREATE THIS (new file)
```

---

## CURRENT SOURCE CODE (read all of this before making changes)

### go.mod
```go
module github.com/Atharv-3105/Graph-Duel   // BUG: wrong name, must be Semantic-Duel

go 1.24.4

require (
    github.com/google/uuid v1.6.0
    github.com/gorilla/websocket v1.5.3
)
```

---

### cmd/server/main.go
```go
package main

import (
    "encoding/json"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/Atharv-3105/Graph-Duel/internal/config"
    "github.com/Atharv-3105/Graph-Duel/internal/logger"
    "github.com/Atharv-3105/Graph-Duel/internal/matchmaker"
    "github.com/Atharv-3105/Graph-Duel/internal/metrics"
    "github.com/Atharv-3105/Graph-Duel/internal/room"
    "github.com/Atharv-3105/Graph-Duel/internal/semantic"
    "github.com/Atharv-3105/Graph-Duel/internal/target"
    "github.com/Atharv-3105/Graph-Duel/internal/ws"
)

func main() {
    cfg := config.Load
    log := logger.New()
    hub := ws.NewHub(log)
    semanticClient := semantic.New(cfg().SemanticURL)
    targetProvider := target.New(target.DefaultWords)
    roomManager := room.NewManager(log)
    cleanupCh := make(chan string, 16)
    matchmaker := matchmaker.New(roomManager, log, semanticClient, targetProvider, cleanupCh, cfg().GameDuration, cfg().RateLimitSeconds)

    go hub.Run()

    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        ws.ServeWS(hub, w, r)
    })

    //temp-hook
    go func() {
        for client := range hub.OnConnect {
            client.SetMessageHandler(func(clientID string, msg ws.ClientMessage) {

                room, ok := roomManager.RoomForClient(clientID)
                if !ok {
                    log.Println("[ROOM] no room for client:", clientID)
                    return
                }

                switch msg.Type {
                case string(ws.ClientWordSubmit):
                    var payload ws.WordSubmitPayload
                    if err := json.Unmarshal(msg.Payload, &payload); err != nil {
                        log.Println("[WS] invalid WORD_SUBMIT payload")
                        return
                    }
                    room.HandleWord(clientID, payload.Word)

                default:
                    log.Println("[WS] unknown message type:", msg.Type)
                }
            })
            matchmaker.Enqueue(client)
        }
    }()

    go func() {
        for client := range hub.OnDisconnect {
            log.Println("[MAIN] client disconnected:", client.ID)
            roomManager.HandleDisconnect(client.ID)
        }
    }()

    go func() {
        for roomID := range cleanupCh {
            log.Println("[MAIN] cleanup received for:", roomID)
            p1, p2 := roomManager.CleanupRoom(roomID)
            if p1 == nil || p2 == nil {
                log.Println("[MAIN] cleanup returned nil players")
                continue
            }

            go func() {
                time.Sleep(2 * time.Second)
                log.Println("[MAIN] requeueing players: ", p1.ID, p2.ID)
                matchmaker.Enqueue(p1)
                matchmaker.Enqueue(p2)
            }()
        }
    }()

    http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(metrics.SnapShot())
    })

    srv := &http.Server{
        Addr: ":" + cfg().ServerPort,
    }

    go func() {
        log.Println("[MAIN] Websocket server started on :8080")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen error: %v", err)
        }
    }()

    shutdown := make(chan os.Signal, 1)
    signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

    <-shutdown
    log.Println("[MAIN] shutdown signal received")
}
```

---

### internal/ws/hub.go
```go
package ws

import "github.com/Atharv-3105/Graph-Duel/internal/logger"

type Hub struct {
    clients      map[*Client]bool
    register     chan *Client
    unregister   chan *Client
    OnConnect    chan *Client    // BUG: unbuffered — blocks Hub.Run() under load
    OnDisconnect chan *Client    // BUG: same issue
    log          *logger.Logger
}

func NewHub(log *logger.Logger) *Hub {
    return &Hub{
        clients:      make(map[*Client]bool),
        register:     make(chan *Client),
        unregister:   make(chan *Client),
        OnConnect:    make(chan *Client),      // BUG: needs buffer
        OnDisconnect: make(chan *Client),      // BUG: needs buffer
        log:          log,
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true
            h.log.Info("[WS] client connected", "active", len(h.clients))
            h.OnConnect <- client             // BUG: blocks if consumer is busy

        case client := <-h.unregister:
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
                h.log.Info("[WS] client disconnected", "active", len(h.clients))
                h.OnDisconnect <- client      // BUG: blocks if consumer is busy
            }
        }
    }
}
```

---

### internal/ws/client.go
```go
package ws

import (
    "encoding/json"
    "log"
    "time"

    "github.com/gorilla/websocket"
)

const (
    writeWait  = 10 * time.Second
    pongWait   = 60 * time.Second
    pingPeriod = pongWait * 9 / 10
)

type MessageHandler func(clientID string, msg ClientMessage)

type Client struct {
    ID   string
    conn *websocket.Conn
    send chan []byte
    onMsg MessageHandler
}

type IncomingMessage struct {   // DEAD CODE — never used
    Type string `json:"type"`
    Word string `json:"word"`
}

func NewClient(conn *websocket.Conn) *Client {
    return &Client{
        conn: conn,
        send: make(chan []byte, 256),
    }
}

func (c *Client) ReadPump(unregister func(*Client)) {
    defer func() {
        unregister(c)
        c.conn.Close()
    }()

    c.conn.SetReadLimit(512)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, msg, err := c.conn.ReadMessage()
        if err != nil {
            return
        }

        var envelope ClientMessage
        if err := json.Unmarshal(msg, &envelope); err != nil {
            log.Println("[WS] invalid message format")
            continue
        }

        if c.onMsg != nil {
            c.onMsg(c.ID, envelope)
        }
    }
}

func (c *Client) WritePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()

    for {
        select {
        case msg, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            c.conn.WriteMessage(websocket.TextMessage, msg)

        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

func (c *Client) SetMessageHandler(h MessageHandler) {
    c.onMsg = h
}

func (c *Client) Send(eventType ServerEventType, payload any) {
    defer func() {
        if r := recover(); r != nil {
            // Ignore sends to closed channels
        }
    }()

    msg := EventMessage{
        Type:    eventType,
        Payload: payload,
    }

    b, err := json.Marshal(msg)
    if err != nil {
        return
    }

    select {
    case c.send <- b:
    default:
    }
}
```

---

### internal/ws/handler.go
```go
package ws

import (
    "net/http"
    "log"
    "github.com/gorilla/websocket"
    "github.com/google/uuid"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
    hub.log.Info("[WS] websocket upgrade request", r.RemoteAddr)
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }

    client := NewClient(conn)
    client.ID = uuid.NewString()
    hub.register <- client

    go client.WritePump()
    go client.ReadPump(func(c *Client) {
        hub.unregister <- c
    })
}

// DEAD CODE — never called, remove it
func (c *Client) handleMessage(m IncomingMessage) {
    log.Println("[WS] message received:", m.Type, m.Word)
}
```

---

### internal/ws/messages.go
```go
package ws

import "encoding/json"

type ClientMessage struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"`
}

type ClientMessageType string

const (
    ClientWordSubmit ClientMessageType = "WORD_SUBMIT"
)

type WordSubmitPayload struct {
    Word string `json:"word"`
}

type EventMessage struct {
    Type    ServerEventType `json:"type"`
    Payload any             `json:"payload"`
}

type ServerEventType string

const (
    EventGameStart   ServerEventType = "GAME_START"
    EventScoreUpdate ServerEventType = "SCORE_UPDATE"
    EventGameOver    ServerEventType = "GAME_OVER"
)

type GameStartPayload struct {
    Target   string `json:"target"`
    Duration int    `json:"duration"`
}

type ScoreUpdatePayload struct {
    PlayerID string `json:"playerId"`
    Score    int    `json:"score"`
}

type GameOverPayload struct {
    Winner string         `json:"winner"`
    Scores map[string]int `json:"scores"`
}
```

---

### internal/room/room.go
```go
package room

import (
    "log"
    "time"
    "unicode"

    "github.com/Atharv-3105/Graph-Duel/internal/game"
    "github.com/Atharv-3105/Graph-Duel/internal/metrics"
    "github.com/Atharv-3105/Graph-Duel/internal/semantic"
    "github.com/Atharv-3105/Graph-Duel/internal/ws"
)

type Room struct {
    ID      string
    Player1 *ws.Client
    Player2 *ws.Client
    State   *GameState

    semantic         *semantic.Client
    lastSubmit       map[string]time.Time

    onCleanup        func(roomID string)
    gameDuration     int
    rateLimitSeconds int   // BUG: stored but never used in the rate limit check
}

func NewRoom(id string, p1, p2 *ws.Client, sc *semantic.Client, onCleanup func(roomID string), gameDuration int, rateLimitSeconds int) *Room {
    return &Room{
        ID:               id,
        Player1:          p1,
        Player2:          p2,
        semantic:         sc,
        lastSubmit:       make(map[string]time.Time),
        onCleanup:        onCleanup,
        gameDuration:     gameDuration,
    }
}

func (r *Room) Start(target string) {
    r.State = &GameState{
        TargetWord: target,
        Scores:     make(map[string]int),
    }
    game.StartGame(&r.State.State)   // BUG: hardcodes 60s, ignores r.gameDuration

    metrics.IncGamesStarted()

    r.broadcast(ws.EventGameStart, ws.GameStartPayload{
        Target:   target,
        Duration: r.gameDuration,
    })

    go func() {
        time.Sleep(time.Until(r.State.EndsAt))
        r.endGame()                  // BUG: data race — runs in separate goroutine,
    }()                              //      no mutex protecting r.State
}

func (r *Room) HandleWord(playerID, word string) {
    // BUG: no mutex — data race with endGame goroutine

    if r.State == nil || r.State.Status != game.Active {
        return
    }

    // RATE LIMITING — BUG: hardcodes time.Second, ignores r.rateLimitSeconds
    now := time.Now()
    last, ok := r.lastSubmit[playerID]
    if ok && now.Sub(last) < time.Second {
        return
    }
    r.lastSubmit[playerID] = now

    // INPUT VALIDATION
    if len(word) == 0 || len(word) > 32 {
        return
    }
    for _, ch := range word {
        if !unicode.IsLetter(ch) {
            return
        }
    }

    // SEMANTIC CALL — BUG: blocks the goroutine for up to 2 seconds
    sim, err := r.semantic.Similarity(word, r.State.TargetWord)
    if err != nil {
        log.Println("[ROOM] semantic error", err)
        return
    }

    score := game.SimilarityToScore(sim)

    prev, exists := r.State.Scores[playerID]
    if !exists || score > prev {
        r.State.Scores[playerID] = score
        r.broadcast(ws.EventScoreUpdate, ws.ScoreUpdatePayload{
            PlayerID: playerID,
            Score:    score,
        })
    }
}

func (r *Room) ForceEnd(leaver string) {
    // BUG: no mutex
    if r.State == nil || r.State.Status == game.Finished {
        return
    }

    r.State.Status = game.Finished

    winner := r.Player1.ID
    if leaver == r.Player1.ID {
        winner = r.Player2.ID
    }

    metrics.IncGamesCompleted()

    r.broadcast(ws.EventGameOver, ws.GameOverPayload{
        Winner: winner,
        Scores: r.State.Scores,
    })

    r.cleanup()
}

func (r *Room) endGame() {
    // BUG: no mutex — concurrent access with HandleWord
    if r.State == nil || r.State.Status == game.Finished {
        return
    }

    r.State.Status = game.Finished

    // BUG: tie-breaking is non-deterministic (Go map iteration is random)
    winner := ""
    high := -1
    for pid, score := range r.State.Scores {
        if score > high {
            high = score
            winner = pid
        }
    }

    metrics.IncGamesCompleted()

    r.broadcast(ws.EventGameOver, ws.GameOverPayload{
        Winner: winner,
        Scores: r.State.Scores,
    })

    r.cleanup()
}

func (r *Room) broadcast(eventType ws.ServerEventType, payload any) {
    r.Player1.Send(eventType, payload)
    r.Player2.Send(eventType, payload)
}

func (r *Room) cleanup() {
    if r.onCleanup != nil {
        r.onCleanup(r.ID)
    }
}
```

---

### internal/room/manager.go
```go
package room

import (
    "sync"
    "github.com/Atharv-3105/Graph-Duel/internal/logger"
    "github.com/Atharv-3105/Graph-Duel/internal/metrics"
    "github.com/Atharv-3105/Graph-Duel/internal/ws"
)

type Manager struct {
    rooms        map[string]*Room
    clientToRoom map[string]*Room
    mu           sync.RWMutex
    log          *logger.Logger
}

func NewManager(log *logger.Logger) *Manager {
    return &Manager{
        rooms:        make(map[string]*Room),
        clientToRoom: make(map[string]*Room),
        log:          log,
    }
}

func (m *Manager) Add(room *Room) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.rooms[room.ID] = room
    m.clientToRoom[room.Player1.ID] = room
    m.clientToRoom[room.Player2.ID] = room
}

func (m *Manager) RoomForClient(clientID string) (*Room, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    room, ok := m.clientToRoom[clientID]
    return room, ok
}

func (m *Manager) HandleDisconnect(clientID string) {
    // BUG: calls room.ForceEnd (which sends to cleanupCh) AND CleanupRoom directly.
    // CleanupRoom gets called twice — once here, once from the cleanupCh consumer in main.go.
    room, ok := m.RoomForClient(clientID)
    if !ok {
        return
    }
    metrics.IncDisconnects()
    room.ForceEnd(clientID)
    m.CleanupRoom(room.ID)  // BUG: redundant — ForceEnd already triggers cleanup via cleanupCh
}

func (m *Manager) CleanupRoom(roomID string) (*ws.Client, *ws.Client) {
    m.mu.Lock()
    defer m.mu.Unlock()

    room, ok := m.rooms[roomID]
    if !ok {
        return nil, nil
    }

    p1 := room.Player1
    p2 := room.Player2

    delete(m.clientToRoom, p1.ID)
    delete(m.clientToRoom, p2.ID)
    delete(m.rooms, roomID)

    return p1, p2
}
```

---

### internal/room/state.go
```go
package room

import "github.com/Atharv-3105/Graph-Duel/internal/game"

type GameState struct {
    TargetWord string
    Scores     map[string]int
    game.State
}
```

---

### internal/room/events.go
```go
package room

// DEAD CODE — EventType and Event are defined but never used in the codebase.
// The room uses ws.ServerEventType directly. This file can be deleted.

type EventType string

const (
    GameStart   EventType = "GAME_START"
    WordSubmit  EventType = "WORD_SUBMIT"
    ScoreUpdate EventType = "SCORE_UPDATE"
    GameOver    EventType = "GAME_OVER"
)

type Event struct {
    Type    EventType   `json:"type"`
    Payload interface{} `json:"payload"`
}
```

---

### internal/game/engine.go
```go
package game

import "time"

type State struct {
    Status Status
    EndsAt time.Time
}

type Status string

const (
    Waiting  Status = "WAITING"
    Active   Status = "ACTIVE"
    Finished Status = "FINISHED"
)

// BUG: hardcodes 60 seconds, ignores the gameDuration config value
func StartGame(state *State) {
    state.Status = Active
    state.EndsAt = time.Now().Add(60 * time.Second)
}

func IsGameOver(state *State) bool {
    return time.Now().After(state.EndsAt)
}
```

---

### internal/game/scoring.go
```go
package game

import "math"

func SimilarityToScore(sim float64) int {
    if sim < 0 { sim = 0 }
    if sim > 1 { sim = 1 }
    return int(math.Round(sim * 100))
}
```

---

### internal/semantic/client.go
```go
package semantic

// THIS FILE WILL BE REPLACED — see instructions below.
// Current implementation makes HTTP calls to a Python FastAPI service.
// Replace with LocalClient that uses pre-computed embeddings.

import (
    "bytes"
    "encoding/json"
    "net/http"
    "time"
)

type Client struct {
    baseURL string
    client  *http.Client
}

func New(baseURL string) *Client {
    return &Client{
        baseURL: baseURL,
        client:  &http.Client{Timeout: 2 * time.Second},
    }
}

type request struct {
    Word   string `json:"word"`
    Target string `json:"target"`
}

type response struct {
    Similarity float64 `json:"similarity"`
}

func (c *Client) Similarity(word, target string) (float64, error) {
    reqBody, _ := json.Marshal(request{Word: word, Target: target})
    resp, err := c.client.Post(c.baseURL+"/similarity", "application/json", bytes.NewBuffer(reqBody))
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()
    var res response
    if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
        return 0, err
    }
    return res.Similarity, nil
}
```

---

### internal/config/config.go
```go
package config

import (
    "log"
    "os"
    "strconv"
)

type Config struct {
    ServerPort       string
    SemanticURL      string
    GameDuration     int
    RateLimitSeconds int
}

func Load() *Config {
    return &Config{
        ServerPort:       getEnv("SERVER_PORT", "8080"),
        SemanticURL:      getEnv("SEMANTIC_SERVICE_URL", "http://localhost:8001"),
        GameDuration:     getEnvInt("GAME_DURATION_SECONDS", 60),
        RateLimitSeconds: getEnvInt("RATE_LIMIT_SECONDS", 1),
    }
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" { return v }
    return fallback
}

func getEnvInt(key string, fallback int) int {
    if v := os.Getenv(key); v != "" {
        i, err := strconv.Atoi(v)
        if err != nil {
            log.Fatalf("[CONFIG] Invalid int env %s = %s", key, v)
        }
        return i
    }
    return fallback
}
```

---

### internal/metrics/metrics.go
```go
package metrics

import "sync/atomic"

var GamesStarted   int64
var GamesCompleted int64
var Disconnects    int64
// BUG: README documents "active_connections" but it is missing from SnapShot()

func IncGamesStarted()   { atomic.AddInt64(&GamesStarted, 1) }
func IncGamesCompleted() { atomic.AddInt64(&GamesCompleted, 1) }
func IncDisconnects()    { atomic.AddInt64(&Disconnects, 1) }

func SnapShot() map[string]int64 {
    return map[string]int64{
        "games_started":   atomic.LoadInt64(&GamesStarted),
        "games_completed": atomic.LoadInt64(&GamesCompleted),
        "disconnects":     atomic.LoadInt64(&Disconnects),
        // missing: "active_connections"
    }
}
```

---

### internal/matchmaker/matchmaker.go
```go
package matchmaker

import (
    "fmt"
    "sync"
    "time"

    "github.com/Atharv-3105/Graph-Duel/internal/logger"
    "github.com/Atharv-3105/Graph-Duel/internal/room"
    "github.com/Atharv-3105/Graph-Duel/internal/semantic"
    "github.com/Atharv-3105/Graph-Duel/internal/target"
    "github.com/Atharv-3105/Graph-Duel/internal/ws"
)

type Matchmaker struct {
    queue           []*ws.Client
    mu              sync.Mutex
    rm              *room.Manager
    log             *logger.Logger
    semantic        *semantic.Client
    targetProvider  *target.Provider
    cleanupCh       chan<- string
    gameDuration    int
    rateLimitSeconds int
}

func New(rm *room.Manager, log *logger.Logger, sc *semantic.Client, targetProvider *target.Provider, cleanupCh chan<- string, gameDuration int, rateLimitSeconds int) *Matchmaker {
    return &Matchmaker{
        queue:            make([]*ws.Client, 0),
        rm:               rm,
        log:              log,
        semantic:         sc,
        targetProvider:   targetProvider,
        cleanupCh:        cleanupCh,
        gameDuration:     gameDuration,
        rateLimitSeconds: rateLimitSeconds,
    }
}

func (m *Matchmaker) Enqueue(client *ws.Client) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.queue = append(m.queue, client)

    if len(m.queue) >= 2 {
        p1 := m.queue[0]
        p2 := m.queue[1]
        m.queue = m.queue[2:]

        roomID := fmt.Sprintf("room-%d", time.Now().UnixNano())

        onCleanup := func(roomID string) {
            m.cleanupCh <- roomID
        }

        room := room.NewRoom(roomID, p1, p2, m.semantic, onCleanup, m.gameDuration, m.rateLimitSeconds)
        targetWord := m.targetProvider.Random()
        room.Start(targetWord)
        m.rm.Add(room)
    }
}
```

---

### internal/logger/logger.go
```go
package logger

import (
    "log"
    "os"
)

type Logger struct {
    *log.Logger
}

func New() *Logger {
    return &Logger{
        Logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
    }
}

func (l *Logger) Info(msg string, fields ...any) {
    l.Println(append([]any{"[INFO]", msg}, fields...)...)
}

func (l *Logger) Error(msg string, fields ...any) {
    l.Println(append([]any{"[ERROR]", msg}, fields...)...)
}
```

---

### internal/target/provider.go
```go
package target

import (
    "math/rand"
    "time"
)

type Provider struct {
    words []string
    rng   *rand.Rand
}

func New(words []string) *Provider {
    return &Provider{
        words: words,
        rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
    }
}

func (p *Provider) Random() string {
    if len(p.words) == 0 {
        return "default"
    }
    return p.words[p.rng.Intn(len(p.words))]
}
```

---

### internal/target/words.go
```go
package target

var DefaultWords = []string{
    "fire", "water", "earth", "wind", "light",
    "dark", "metal", "nature", "energy", "time",
}
```

---

## TASK — WHAT I WANT YOU TO DO

### Part 1: Fix All Bugs

Fix every bug annotated with `// BUG:` in the code above. Here is a complete list:

#### Bug 1 — Wrong module name in go.mod
Rename `github.com/Atharv-3105/Graph-Duel` to `github.com/Atharv-3105/Semantic-Duel` in `go.mod` AND update every import path in every `.go` file accordingly.

#### Bug 2 — Hub deadlock: unbuffered OnConnect / OnDisconnect channels
In `ws/hub.go`, `NewHub()`: change `OnConnect` and `OnDisconnect` to buffered channels with size 64.

#### Bug 3 — Data race on Room state
In `room/room.go`: add a `sync.RWMutex` field named `mu` to the `Room` struct. Lock it (`mu.Lock()`) at the top of `HandleWord`, `endGame`, and `ForceEnd`. Use `RLock` if only reading.

#### Bug 4 — Semantic HTTP call blocks the dispatch goroutine
In `room/room.go`, inside `HandleWord`: after all validation passes, dispatch the similarity lookup + score update inside a `go func()`. Make sure the goroutine re-acquires the mutex and re-checks `r.State.Status != game.Active` before writing (the game might have ended during the HTTP call).

#### Bug 5 — gameDuration is ignored in StartGame
In `game/engine.go`: change `StartGame(state *State)` to `StartGame(state *State, durationSeconds int)` and use `time.Duration(durationSeconds) * time.Second` instead of the hardcoded `60 * time.Second`.
Update the call site in `room/room.go` `Start()` to pass `r.gameDuration`.

#### Bug 6 — rateLimitSeconds is stored but not used
In `room/room.go`, `HandleWord()`: change `time.Second` in the rate limit check to `time.Duration(r.rateLimitSeconds) * time.Second`. Also fix `NewRoom` to actually assign `rateLimitSeconds` to the struct field (it is missing from the return value).

#### Bug 7 — Double cleanup on disconnect
In `room/manager.go`, `HandleDisconnect()`: remove the direct call to `m.CleanupRoom(room.ID)`. Cleanup should flow exclusively through `cleanupCh` → `main.go` → `CleanupRoom`. This avoids the double-cleanup race.

#### Bug 8 — Non-deterministic tie-breaking
In `room/room.go`, `endGame()`: replace the map-iteration winner logic with explicit comparison:
```go
p1Score := r.State.Scores[r.Player1.ID]
p2Score := r.State.Scores[r.Player2.ID]
winner := ""
if p1Score > p2Score {
    winner = r.Player1.ID
} else if p2Score > p1Score {
    winner = r.Player2.ID
}
// empty string means draw
```

#### Bug 9 — active_connections missing from metrics
In `metrics/metrics.go`: add `ActiveConnections int64`, `IncConnections()`, `DecConnections()` atomic helpers and include `"active_connections"` in `SnapShot()`.
In `ws/hub.go` `Run()`: call `metrics.IncConnections()` on register, `metrics.DecConnections()` on unregister.

#### Bug 10 — Dead code cleanup
- Delete `room/events.go` entirely (unused `EventType` and `Event`)
- In `ws/handler.go`: delete the `handleMessage` method and the import of `"log"` if it becomes unused
- In `ws/client.go`: delete the `IncomingMessage` struct

#### Bug 11 — Graceful shutdown in main.go
After `<-shutdown`, add a proper graceful shutdown:
```go
ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
defer cancel()
if err := srv.Shutdown(ctx); err != nil {
    log.Printf("[MAIN] shutdown error: %v", err)
}
log.Println("[MAIN] server stopped cleanly")
```

#### Bug 12 — Add /health endpoint in main.go
```go
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"status":"ok"}`))
})
```

---

### Part 2: Replace the Python ML Service with Pre-computed Embeddings

**Goal:** Eliminate the need to run the Python FastAPI service at runtime. Instead:
1. A Python script (`scripts/precompute.py`) generates word embeddings once and saves them to disk
2. A Go `LocalClient` (`internal/semantic/local.go`) loads those embeddings at startup and answers similarity queries with a pure in-memory cosine similarity — no network calls, no Python at runtime

#### Step A — Create `scripts/precompute.py`

This script:
- Uses `sentence-transformers` with `all-MiniLM-L6-v2` (free, CPU-runnable)
- Reads a built-in vocabulary of at least 3000 common English words
- Embeds all words in batches
- Saves two files: `data/vocabulary.json` (word list) and `data/embeddings.npy` (float32 numpy array, shape [N, 384])
- Prints progress and completion

Include a hardcoded vocabulary of at least 3000 common English words spanning categories: nature, emotions, actions, objects, places, science, food, body, relationships, time, movement, abstract concepts. The words should be lowercase single words only.

Usage: `python scripts/precompute.py`
Output: `data/vocabulary.json` and `data/embeddings.npy`

Requirements (save as `scripts/requirements.txt`):
```
sentence-transformers>=2.2.0
numpy>=1.24.0
```

#### Step B — Create `internal/semantic/local.go`

Replace `internal/semantic/client.go` with `internal/semantic/local.go` that:
- Defines a `LocalClient` struct with fields: `vocab []string`, `embeddings [][]float32`, `index map[string]int`
- Has a constructor `func NewLocal(vocabPath, embeddingsPath string) (*LocalClient, error)` that:
  - Reads `vocabulary.json` into `vocab`
  - Reads `embeddings.npy` into `embeddings` — parse the numpy `.npy` format natively in Go (no external library: read the magic string, header, and raw float32 bytes)
  - Builds the `index` map
- Has a method `Similarity(ctx context.Context, word, target string) (float64, error)` that:
  - Looks up both words in the index
  - If either word is not found, returns `0.0, nil` (unknown word = zero score)
  - Computes cosine similarity in pure Go
  - Returns the score

Also update `internal/semantic/client.go` to keep the old HTTP `Client` but rename it to make the interface clear, OR define a `Semantic` interface:
```go
type Semantic interface {
    Similarity(ctx context.Context, word, target string) (float64, error)
}
```

Both `Client` (HTTP) and `LocalClient` (embedded) implement this interface.

#### Step C — Update all callers

Everywhere `*semantic.Client` is used as a type (in `room/room.go`, `matchmaker/matchmaker.go`), change it to the `semantic.Semantic` interface so both implementations work.

#### Step D — Update main.go to use LocalClient by default

In `main.go`, load the local client:
```go
semanticClient, err := semantic.NewLocal("data/vocabulary.json", "data/embeddings.npy")
if err != nil {
    log.Fatalf("[MAIN] failed to load embeddings: %v", err)
}
```

Remove `SemanticURL` from the config or keep it as a fallback (your choice — just make the default be the local client).

Also remove the `SemanticURL` field from `Config` if it's no longer needed, or make it optional with a comment explaining it's for the legacy HTTP mode.

#### Step E — Update the README section

Replace the Python service startup instructions with:
```
# One-time setup (run once on any machine with Python):
pip install -r scripts/requirements.txt
python scripts/precompute.py
# This creates data/vocabulary.json and data/embeddings.npy

# Then run the Go server (no Python needed at runtime):
go run ./cmd/server
```

---

## OUTPUT FORMAT

Please give me the **complete, final content** of every file that changes. Do not give diffs or partial snippets — give the full file for each changed file, clearly labeled with its path. I will copy-paste them directly.

Files to output:
- `go.mod`
- `cmd/server/main.go`
- `internal/ws/hub.go`
- `internal/ws/client.go`
- `internal/ws/handler.go`  (or confirm deleted sections)
- `internal/ws/messages.go` (if changed)
- `internal/game/engine.go`
- `internal/room/room.go`
- `internal/room/manager.go`
- `internal/room/state.go`
- `internal/room/events.go` → DELETE THIS FILE (confirm)
- `internal/semantic/local.go` ← NEW FILE
- `internal/semantic/client.go` ← updated with Semantic interface
- `internal/config/config.go`
- `internal/metrics/metrics.go`
- `internal/matchmaker/matchmaker.go`
- `scripts/precompute.py` ← NEW FILE
- `scripts/requirements.txt` ← NEW FILE

Do not change any frontend files. Do not change `internal/target/`, `internal/logger/`, or any test files.

After all files, add a section called **"How to Run After These Changes"** with the exact commands in order.
