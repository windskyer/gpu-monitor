package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/windskyer/gpu-monitor/internal/model"
	"github.com/windskyer/gpu-monitor/internal/store"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
}

// Hub manages WebSocket connections and broadcasts snapshots.
type Hub struct {
	mu    sync.Mutex
	conns map[*websocket.Conn]struct{}
}

func newHub() *Hub {
	return &Hub{conns: make(map[*websocket.Conn]struct{})}
}

func (h *Hub) add(c *websocket.Conn) {
	h.mu.Lock()
	h.conns[c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) remove(c *websocket.Conn) {
	h.mu.Lock()
	delete(h.conns, c)
	h.mu.Unlock()
}

func (h *Hub) broadcast(data []byte) {
	h.mu.Lock()
	dead := []*websocket.Conn{}
	for c := range h.conns {
		c.SetWriteDeadline(time.Now().Add(2 * time.Second))
		if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
			dead = append(dead, c)
		}
	}
	h.mu.Unlock()
	for _, c := range dead {
		c.Close()
		h.remove(c)
	}
}

// Server is the HTTP + WebSocket server.
type Server struct {
	ring  *store.Ring
	hub   *Hub
	token string
}

func New(ring *store.Ring, token string) *Server {
	return &Server{ring: ring, hub: newHub(), token: token}
}

// Listener is the store.Listener that feeds the hub.
func (s *Server) Listener(snap *model.Snapshot) {
	data, err := json.Marshal(snap)
	if err != nil {
		log.Printf("[server] marshal: %v", err)
		return
	}
	s.hub.broadcast(data)
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWS)
	mux.HandleFunc("/api/snapshot", s.handleSnapshot)
	mux.Handle("/", http.FileServer(http.FS(staticFS)))
	return mux
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	if s.token != "" && s.token != "change-me" {
		if r.URL.Query().Get("token") != s.token {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[server] ws upgrade: %v", err)
		return
	}
	s.hub.add(conn)

	// Send current snapshot immediately on connect
	if latest := s.ring.Latest(); latest != nil {
		data, _ := json.Marshal(latest)
		conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		conn.WriteMessage(websocket.TextMessage, data)
	}

	// Read loop — just drain to detect disconnects
	go func() {
		defer func() {
			conn.Close()
			s.hub.remove(conn)
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}

func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	snap := s.ring.Latest()
	if snap == nil {
		http.Error(w, "no data", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snap)
}
