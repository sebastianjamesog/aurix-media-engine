package gateway

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ClientSession represents a connected WebSocket client (e.g. Discord bot).
type ClientSession struct {
	SessionID  string
	UserID     string
	ClientName string
	ws         *websocket.Conn
	sendChan   chan []byte
	stopChan   chan struct{}
	mu         sync.Mutex
}

// WSManager manages active client WebSocket sessions.
type WSManager struct {
	sessions map[string]*ClientSession
	mu       sync.RWMutex
}

// NewWSManager creates a client WebSocket session manager.
func NewWSManager() *WSManager {
	return &WSManager{
		sessions: make(map[string]*ClientSession),
	}
}

func generateSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// HandleWebSocket upgrades HTTP requests to WebSocket connections.
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if auth != s.password {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	userID := r.Header.Get("User-Id")
	clientName := r.Header.Get("Client-Name")
	existingSessionID := r.Header.Get("Session-Id")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Upgrade error: %v", err)
		return
	}

	sessionID := existingSessionID
	resumed := false
	if sessionID == "" {
		sessionID = generateSessionID()
	} else {
		resumed = true
	}

	client := &ClientSession{
		SessionID:  sessionID,
		UserID:     userID,
		ClientName: clientName,
		ws:         conn,
		sendChan:   make(chan []byte, 256),
		stopChan:   make(chan struct{}),
	}

	log.Printf("[WS] Client connected: User-Id=%s, Client-Name=%s, Session-Id=%s (resumed=%v)", userID, clientName, sessionID, resumed)

	// Send `ready` payload
	readyPayload := map[string]any{
		"op":        "ready",
		"resumed":   resumed,
		"sessionId": sessionID,
	}
	readyBytes, _ := json.Marshal(readyPayload)
	_ = conn.WriteMessage(websocket.TextMessage, readyBytes)

	go client.writeLoop()
	go client.readLoop()
}

func (c *ClientSession) writeLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case msg := <-c.sendChan:
			c.mu.Lock()
			_ = c.ws.WriteMessage(websocket.TextMessage, msg)
			c.mu.Unlock()
		case <-ticker.C:
			c.mu.Lock()
			_ = c.ws.WriteMessage(websocket.PingMessage, nil)
			c.mu.Unlock()
		}
	}
}

func (c *ClientSession) readLoop() {
	defer func() {
		close(c.stopChan)
		_ = c.ws.Close()
	}()

	for {
		_, _, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

// SendEvent serializes and dispatches an event object to the client session.
func (c *ClientSession) SendEvent(event any) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	select {
	case c.sendChan <- data:
		return nil
	default:
		log.Printf("[WS] Send buffer full for session %s, dropping event", c.SessionID)
		return nil
	}
}
