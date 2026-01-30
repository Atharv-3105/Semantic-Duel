package main

import (
	// "log"
	"encoding/json"
	"net/http"
	"time"
	"os"
	"os/signal"
	"syscall"

	// "time"

	"github.com/Atharv-3105/Graph-Duel/internal/config"
	"github.com/Atharv-3105/Graph-Duel/internal/logger"
	"github.com/Atharv-3105/Graph-Duel/internal/matchmaker"
	"github.com/Atharv-3105/Graph-Duel/internal/metrics"
	"github.com/Atharv-3105/Graph-Duel/internal/room"
	"github.com/Atharv-3105/Graph-Duel/internal/semantic"
	"github.com/Atharv-3105/Graph-Duel/internal/ws"
)


func main() {
	cfg := config.Load
	log := logger.New()
	hub := ws.NewHub(log)
	semanticClient := semantic.New(cfg().SemanticURL)
	roomManager := room.NewManager(log)
	cleanupCh := make(chan string, 16)
	matchmaker := matchmaker.New(roomManager, log, semanticClient, cleanupCh, cfg().GameDuration, cfg().RateLimitSeconds)

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
				log.Println("[MAIN] cleanup returened nil players")
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

	//Metrics END-POINT
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

	//SHUT-DOWN SIGNALS
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	<-shutdown
	log.Println("[MAIN] shutdown signal received")
	
}