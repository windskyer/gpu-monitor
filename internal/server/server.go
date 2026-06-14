package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
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
	mux.HandleFunc("/gpu/ws", s.handleWS)
	mux.HandleFunc("/gpu/api/snapshot", s.handleSnapshot)
	mux.HandleFunc("/gpu", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/gpu/", http.StatusMovedPermanently)
	})
	mux.Handle("/gpu/", http.StripPrefix("/gpu/", http.FileServer(http.FS(staticFS))))
	return s.withMiddleware(mux)
}

// withMiddleware wraps handler with logging and optional token auth.
func (s *Server) withMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// logging: capture status
		lrw := &loggingResponseWriter{ResponseWriter: w, status: 200}
		h.ServeHTTP(lrw, r)
		dur := time.Since(start)
		log.Printf("[http] %s %s %s %d %dB %v", r.RemoteAddr, r.Method, r.URL.Path, lrw.status, lrw.written, dur)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status  int
	written int
}

func (l *loggingResponseWriter) WriteHeader(code int) {
	l.status = code
	l.ResponseWriter.WriteHeader(code)
}

func (l *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := l.ResponseWriter.Write(b)
	l.written += n
	return n, err
}

func (l *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := l.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("ResponseWriter does not support Hijacker")
	}
	return h.Hijack()
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	if s.token != "" && r.URL.Query().Get("token") != s.token {
		log.Printf("[ws] auth failed from %s", r.RemoteAddr)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ws] upgrade error from %s: %v", r.RemoteAddr, err)
		return
	}

	s.hub.add(conn)
	log.Printf("[ws] connected: %s", r.RemoteAddr)

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
			log.Printf("[ws] disconnected: %s", r.RemoteAddr)
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
