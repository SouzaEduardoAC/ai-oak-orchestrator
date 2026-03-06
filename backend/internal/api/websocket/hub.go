package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/ecoza/ai-oak-orchestrator/internal/infrastructure/valkey"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type AgentRunner interface {
	Run(ctx context.Context, session *domain.Session, output chan<- domain.AgentEvent, input <-chan domain.AgentCommand) error
}

// sessionState holds everything needed to route events/commands for one active agent session.
type sessionState struct {
	input    chan domain.AgentCommand
	mu       sync.Mutex          // serializes writes to the current conn
	conn     *websocket.Conn     // current active WebSocket connection (may change on reconnect)
}

func (s *sessionState) write(msgType int, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn == nil {
		return nil // no active connection, drop the write
	}
	// Apply a write deadline so a stalled browser never blocks the agent.
	s.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	err := s.conn.WriteMessage(msgType, data)
	s.conn.SetWriteDeadline(time.Time{}) // clear deadline after write
	return err
}

func (s *sessionState) setConn(c *websocket.Conn) {
	s.mu.Lock()
	s.conn = c
	s.mu.Unlock()
}

type Hub struct {
	clients       map[*websocket.Conn]bool
	broadcast     chan []byte
	register      chan *websocket.Conn
	unregister    chan *websocket.Conn
	sessions      map[string]*sessionState // sessionID -> active agent state
	connSession   map[*websocket.Conn]string // conn -> sessionID
	agent         AgentRunner
	valkey        *valkey.Client
	logger        *zap.Logger
	mu            sync.Mutex
}

func NewHub(logger *zap.Logger, agent AgentRunner, vdb *valkey.Client) *Hub {
	return &Hub{
		clients:     make(map[*websocket.Conn]bool),
		broadcast:   make(chan []byte),
		register:    make(chan *websocket.Conn),
		unregister:  make(chan *websocket.Conn),
		sessions:    make(map[string]*sessionState),
		connSession: make(map[*websocket.Conn]string),
		agent:       agent,
		valkey:      vdb,
		logger:      logger,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Debug("Client registered")
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			// Detach conn from session so writes are dropped until reconnect
			if sid, ok := h.connSession[client]; ok {
				if st, ok := h.sessions[sid]; ok {
					st.mu.Lock()
					if st.conn == client {
						st.conn = nil
					}
					st.mu.Unlock()
				}
				delete(h.connSession, client)
			}
			h.mu.Unlock()
			h.logger.Debug("Client unregistered")
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					h.logger.Error("Write error", zap.Error(err))
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func (h *Hub) HandleWebSocket(c echo.Context) error {
	h.logger.Info("HandleWebSocket called, upgrading connection")
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.logger.Error("WebSocket upgrade failed", zap.Error(err))
		return err
	}
	h.logger.Info("WebSocket upgraded successfully")
	h.register <- ws

	defer func() {
		h.unregister <- ws
	}()

	// Keepalive: ping every 30s, close if no pong within 70s.
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(70 * time.Second))
		return nil
	})
	ws.SetReadDeadline(time.Now().Add(70 * time.Second))

	stopPing := make(chan struct{})
	defer close(stopPing)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// WriteControl is safe to call concurrently with other write methods.
				if err := ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
					return
				}
			case <-stopPing:
				return
			}
		}
	}()

	sessionID := uuid.New().String()
	if user, ok := c.Get("user").(jwt.MapClaims); ok {
		if sub, ok := user["sub"].(string); ok {
			sessionID = sub
		}
	}

	// Attach this conn to the session. If an agent is already running for this
	// session (e.g. reconnect after page refresh), re-use it.
	h.mu.Lock()
	h.connSession[ws] = sessionID
	if st, ok := h.sessions[sessionID]; ok {
		st.setConn(ws)
		h.logger.Info("Reconnected to existing agent session", zap.String("session_id", sessionID))
	}
	h.mu.Unlock()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			h.logger.Error("Failed to unmarshal websocket message", zap.Error(err), zap.String("raw", string(msg)))
			continue
		}

		h.logger.Info("Received websocket message", zap.String("type", wsMsg.Type))

		switch wsMsg.Type {
		case "message":
			h.mu.Lock()
			_, hasRunning := h.sessions[sessionID]
			h.mu.Unlock()
			if !hasRunning {
				go h.runAgent(ws, wsMsg.Payload, sessionID)
			}
		case "approval":
			var approval struct {
				CallID   string `json:"callId"`
				Approved bool   `json:"approved"`
			}
			if err := json.Unmarshal(wsMsg.Payload, &approval); err == nil {
				cmdType := domain.CommandReject
				if approval.Approved {
					cmdType = domain.CommandApprove
				}
				h.mu.Lock()
				if st, ok := h.sessions[sessionID]; ok {
					st.input <- domain.AgentCommand{
						Type:    cmdType,
						Payload: wsMsg.Payload,
					}
				}
				h.mu.Unlock()
			}
		case string(domain.CommandSetModel):
			h.mu.Lock()
			if st, ok := h.sessions[sessionID]; ok {
				st.input <- domain.AgentCommand{
					Type:    domain.AgentCommandType(wsMsg.Type),
					Payload: wsMsg.Payload,
				}
			}
			h.mu.Unlock()
		}
	}

	return nil
}

func (h *Hub) runAgent(conn *websocket.Conn, payload json.RawMessage, sessionID string) {
	ctx := context.Background()
	output := make(chan domain.AgentEvent, 64)
	input := make(chan domain.AgentCommand, 1)

	st := &sessionState{input: input, conn: conn}

	h.mu.Lock()
	h.sessions[sessionID] = st
	h.mu.Unlock()

	session, err := h.valkey.GetSession(ctx, sessionID)
	if err != nil {
		h.logger.Info("Starting new session", zap.String("id", sessionID))
		session = &domain.Session{ID: sessionID}
	}

	session.Messages = append(session.Messages, domain.Message{Role: "user", Content: string(payload)})

	defer func() {
		h.mu.Lock()
		delete(h.sessions, sessionID)
		h.mu.Unlock()
		close(output)

		if err := h.valkey.SaveSession(ctx, session); err != nil {
			h.logger.Error("Failed to save session", zap.Error(err))
		}
	}()

	go func() {
		for event := range output {
			resp, _ := json.Marshal(event)
			st.write(websocket.TextMessage, resp)
		}
	}()

	if err := h.agent.Run(ctx, session, output, input); err != nil {
		h.logger.Error("Agent run error", zap.Error(err))
	}
}
