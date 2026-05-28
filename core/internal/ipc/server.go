package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/google/uuid"
	schema "flint/core/pkg/schema"
)

// Envelope wraps any IPC message with a type discriminator.
type Envelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Handler is called by the server when a message arrives from a client.
type Handler func(clientID string, env Envelope)

// Server listens on a Unix domain socket and multiplexes messages to clients.
type Server struct {
	socketPath string
	handler    Handler
	listener   net.Listener

	mu      sync.RWMutex
	clients map[string]*client
}

type client struct {
	id   string
	conn net.Conn
	enc  *json.Encoder
}

// NewServer creates a server that will listen on socketPath.
func NewServer(socketPath string, handler Handler) *Server {
	return &Server{
		socketPath: socketPath,
		handler:    handler,
		clients:    make(map[string]*client),
	}
}

// Listen starts the Unix socket listener. Blocks until Close is called.
func (s *Server) Listen() error {
	// Remove stale socket
	_ = removeSocket(s.socketPath)

	l, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("ipc: listen %s: %w", s.socketPath, err)
	}
	s.listener = l

	for {
		conn, err := l.Accept()
		if err != nil {
			return nil // closed
		}
		go s.handleConn(conn)
	}
}

// Close shuts down the listener and disconnects all clients.
func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// Broadcast sends a typed message to all connected clients.
func (s *Server) Broadcast(msgType string, payload any) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return
	}
	env := Envelope{Type: msgType, Payload: raw}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.clients {
		_ = c.enc.Encode(env)
	}
}

// BroadcastObservation is a typed helper for observation messages.
func (s *Server) BroadcastObservation(obs schema.Observation) {
	s.Broadcast("observation", obs)
}

// BroadcastSessionState is a typed helper for session state messages.
func (s *Server) BroadcastSessionState(ss schema.SessionState) {
	s.Broadcast("sessionState", ss)
}

// BroadcastTripwireFired is a typed helper for tripwire events.
func (s *Server) BroadcastTripwireFired(tf schema.TripwireFired) {
	s.Broadcast("tripwireFired", tf)
}

// BroadcastAutoFix is a typed helper for auto-fix events.
func (s *Server) BroadcastAutoFix(af schema.AutoFix) {
	s.Broadcast("autoFix", af)
}

// ---------------------------------------------------------------------------
// Internal
// ---------------------------------------------------------------------------

func (s *Server) handleConn(conn net.Conn) {
	// Unix sockets return "" from RemoteAddr — use a stable UUID instead.
	id := uuid.New().String()

	c := &client{id: id, conn: conn, enc: json.NewEncoder(conn)}

	s.mu.Lock()
	s.clients[id] = c
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, id)
		s.mu.Unlock()
		conn.Close()
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var env Envelope
		if err := json.Unmarshal(line, &env); err != nil {
			continue
		}
		if s.handler != nil {
			s.handler(id, env)
		}
	}
}
