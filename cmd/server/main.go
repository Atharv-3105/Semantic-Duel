package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Atharv-3105/Semantic-Duel/internal/config"
	"github.com/Atharv-3105/Semantic-Duel/internal/logger"
	"github.com/Atharv-3105/Semantic-Duel/internal/matchmaker"
	"github.com/Atharv-3105/Semantic-Duel/internal/metrics"
	"github.com/Atharv-3105/Semantic-Duel/internal/room"
	"github.com/Atharv-3105/Semantic-Duel/internal/semantic"
	"github.com/Atharv-3105/Semantic-Duel/internal/target"
	"github.com/Atharv-3105/Semantic-Duel/internal/ws"
)

func main() {
	cfg := config.Load()
	log := logger.New()

	hub := ws.NewHub(log)
	semanticClient := semantic.New(cfg.SemanticURL)
	targetProvider := target.New(target.DefaultWords)
	roomManager := room.NewManager(log)
	cleanupCh := make(chan string, 16)
	mm := matchmaker.New(roomManager, log, semanticClient, targetProvider, cleanupCh, cfg.GameDuration, cfg.RateLimitSeconds)

	go hub.Run()

	// Route WebSocket connections.
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWS(hub, w, r)
	})

	// Health check — required for Docker, load balancers, and uptime monitors.
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Metrics snapshot — inspect counters at runtime.
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics.SnapShot())
	})

	// When a client connects: register its message handler, then enqueue for matchmaking.
	go func() {
		for client := range hub.OnConnect {
			client.SetMessageHandler(func(clientID string, msg ws.ClientMessage) {
				r, ok := roomManager.RoomForClient(clientID)
				if !ok {
					log.Println("[MAIN] no room for client:", clientID)
					return
				}

				switch msg.Type {
				case string(ws.ClientWordSubmit):
					var payload ws.WordSubmitPayload
					if err := json.Unmarshal(msg.Payload, &payload); err != nil {
						log.Println("[WS] invalid WORD_SUBMIT payload")
						return
					}
					r.HandleWord(clientID, payload.Word)

				default:
					log.Println("[WS] unknown message type:", msg.Type)
				}
			})
			mm.Enqueue(client)
		}
	}()

	// When a client disconnects: force the game to end.
	go func() {
		for client := range hub.OnDisconnect {
			log.Println("[MAIN] client disconnected:", client.ID)
			roomManager.HandleDisconnect(client.ID)
		}
	}()

	// When a game ends: clean up the room and re-queue both players after a short delay.
	go func() {
		for roomID := range cleanupCh {
			log.Println("[MAIN] cleanup for room:", roomID)
			p1, p2 := roomManager.CleanupRoom(roomID)
			if p1 == nil || p2 == nil {
				log.Println("[MAIN] cleanup returned nil players — room already cleaned")
				continue
			}

			go func() {
				time.Sleep(2 * time.Second)
				log.Println("[MAIN] re-queuing players:", p1.ID, p2.ID)
				mm.Enqueue(p1)
				mm.Enqueue(p2)
			}()
		}
	}()

	srv := &http.Server{
		Addr: ":" + cfg.ServerPort,
	}

	go func() {
		log.Println("[MAIN] server listening on :" + cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[MAIN] listen error: %v", err)
		}
	}()

	// Block until SIGINT or SIGTERM.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	log.Println("[MAIN] shutdown signal received — draining active connections...")

	// Give in-flight requests up to 15 seconds to complete before forcing exit.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[MAIN] graceful shutdown error: %v", err)
	}

	log.Println("[MAIN] server stopped cleanly")
}