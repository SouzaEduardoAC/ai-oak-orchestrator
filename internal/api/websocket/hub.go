package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/ecoza/ai-oak-orchestrator/internal/infrastructure/valkey"
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

type Hub struct {
	clients         map[*websocket.Conn]bool
	broadcast       chan []byte
	register        chan *websocket.Conn
	unregister      chan *websocket.Conn
	pendingCommands map[*websocket.Conn]chan domain.AgentCommand
	agent           AgentRunner
	valkey          *valkey.Client
	logger          *zap.Logger
	mu              sync.Mutex
}

func NewHub(logger *zap.Logger, agent AgentRunner, vdb *valkey.Client) *Hub {
	return &Hub{
		clients:         make(map[*websocket.Conn]bool),
		broadcast:       make(chan []byte),
		register:        make(chan *websocket.Conn),
		unregister:      make(chan *websocket.Conn),
		pendingCommands: make(map[*websocket.Conn]chan domain.AgentCommand),
		agent:           agent,
		valkey:          vdb,
		logger:          logger,
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
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	h.register <- ws

	defer func() {
		h.unregister <- ws
	}()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			continue
		}

		switch wsMsg.Type {
		case "message":
			go h.runAgent(ws, wsMsg.Payload)
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
				if ch, ok := h.pendingCommands[ws]; ok {
					ch <- domain.AgentCommand{
						Type:    cmdType,
						Payload: wsMsg.Payload,
					}
				}
				h.mu.Unlock()
			}
		case string(domain.CommandSetModel):
			h.mu.Lock()
			if ch, ok := h.pendingCommands[ws]; ok {
				ch <- domain.AgentCommand{
					Type:    domain.AgentCommandType(wsMsg.Type),
					Payload: wsMsg.Payload,
				}
			}
			h.mu.Unlock()
		}
	}
	
	return nil
}

func (h *Hub) runAgent(conn *websocket.Conn, payload json.RawMessage) {
	ctx := context.Background()
	output := make(chan domain.AgentEvent)
	input := make(chan domain.AgentCommand, 1)

	h.mu.Lock()
	h.pendingCommands[conn] = input
	h.mu.Unlock()

	sessionID := "default-session"

	session, err := h.valkey.GetSession(ctx, sessionID)
	if err != nil {
		h.logger.Info("Starting new session", zap.String("id", sessionID))
		session = &domain.Session{ID: sessionID}
	}

	session.Messages = append(session.Messages, domain.Message{Role: "user", Content: string(payload)})

	defer func() {
		h.mu.Lock()
		delete(h.pendingCommands, conn)
		h.mu.Unlock()
		close(output)
		
		if err := h.valkey.SaveSession(ctx, session); err != nil {
			h.logger.Error("Failed to save session", zap.Error(err))
		}
	}()

	go func() {
		for event := range output {
			resp, _ := json.Marshal(event)
			conn.WriteMessage(websocket.TextMessage, resp)
		}
	}()

	if err := h.agent.Run(ctx, session, output, input); err != nil {
		h.logger.Error("Agent run error", zap.Error(err))
	}
}
