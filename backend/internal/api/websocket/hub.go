package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
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
	Run(ctx context.Context, session *domain.Session, output chan<- string) error
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	agent      AgentRunner
	logger     *zap.Logger
	mu         sync.Mutex
}

func NewHub(logger *zap.Logger, agent AgentRunner) *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		agent:      agent,
		logger:     logger,
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

// ... existing Run method ...

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

		if wsMsg.Type == "chat" {
			// Start agent run
			go h.runAgent(ws, wsMsg.Payload)
		}
	}
	
	return nil
}

func (h *Hub) runAgent(conn *websocket.Conn, payload json.RawMessage) {
	ctx := context.Background()
	output := make(chan string)
	
	// Create a dummy session for now
	session := &domain.Session{
		ID: "temp",
		Messages: []domain.Message{
			{Role: "user", Content: string(payload)},
		},
	}

	go func() {
		for chunk := range output {
			resp, _ := json.Marshal(map[string]interface{}{
				"event": "agent:token",
				"data":  chunk,
			})
			conn.WriteMessage(websocket.TextMessage, resp)
		}
	}()

	if err := h.agent.Run(ctx, session, output); err != nil {
		h.logger.Error("Agent run error", zap.Error(err))
	}
	close(output)
}
