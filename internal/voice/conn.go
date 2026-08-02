package voice

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// VoiceState represents the authentication parameters required for Discord voice connection.
type VoiceState struct {
	GuildID   string `json:"guild_id"`
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Token     string `json:"token"`
	Endpoint  string `json:"endpoint"`
}

// Conn manages a Discord Voice Gateway WebSocket connection.
type Conn struct {
	state      VoiceState
	ws         *websocket.Conn
	udp        *UDPConn
	ssrc       uint32
	ip         string
	port       uint16
	secretKey  [32]byte
	modes      []string
	heartbeat  time.Duration
	stopChan   chan struct{}
	mu         sync.Mutex
}

// NewConn creates a new Discord Voice connection manager.
func NewConn(state VoiceState) *Conn {
	return &Conn{
		state:    state,
		stopChan: make(chan struct{}),
	}
}

// Connect opens the Voice WebSocket Gateway and performs the Identify & Select Protocol handshakes.
func (c *Conn) Connect(ctx context.Context) error {
	wsURL := fmt.Sprintf("wss://%s?v=8", c.state.Endpoint)
	log.Printf("[Voice] Connecting to Discord Voice Gateway: %s", wsURL)

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return fmt.Errorf("voice websocket dial failed: %w", err)
	}
	c.ws = conn

	// Read Hello message (Opcode 8)
	var hello struct {
		Op int `json:"op"`
		D  struct {
			HeartbeatInterval float64 `json:"heartbeat_interval"`
		} `json:"d"`
	}

	if err := c.ws.ReadJSON(&hello); err != nil {
		return fmt.Errorf("failed to read voice hello: %w", err)
	}

	c.heartbeat = time.Duration(hello.D.HeartbeatInterval) * time.Millisecond
	go c.heartbeatLoop()

	// Send Opcode 0 (Identify)
	identifyPayload := map[string]any{
		"op": 0,
		"d": map[string]any{
			"server_id":  c.state.GuildID,
			"user_id":    c.state.UserID,
			"session_id": c.state.SessionID,
			"token":      c.state.Token,
		},
	}

	if err := c.ws.WriteJSON(identifyPayload); err != nil {
		return fmt.Errorf("failed to send identify payload: %w", err)
	}

	// Read Opcode 2 (Ready)
	var ready struct {
		Op int `json:"op"`
		D  struct {
			SSRC  uint32   `json:"ssrc"`
			IP    string   `json:"ip"`
			Port  uint16   `json:"port"`
			Modes []string `json:"modes"`
		} `json:"d"`
	}

	if err := c.ws.ReadJSON(&ready); err != nil {
		return fmt.Errorf("failed to read voice ready: %w", err)
	}

	c.ssrc = ready.D.SSRC
	c.ip = ready.D.IP
	c.port = ready.D.Port
	c.modes = ready.D.Modes

	// Connect UDP Socket & Perform STUN IP Discovery
	udpAddr := fmt.Sprintf("%s:%d", c.ip, c.port)
	udpConn, err := NewUDPConn(udpAddr, c.ssrc)
	if err != nil {
		return fmt.Errorf("failed to establish UDP socket: %w", err)
	}
	c.udp = udpConn

	pubIP, pubPort, err := c.udp.DiscoverIP()
	if err != nil {
		return fmt.Errorf("STUN IP discovery failed: %w", err)
	}

	// Send Opcode 1 (Select Protocol)
	selectProtocol := map[string]any{
		"op": 1,
		"d": map[string]any{
			"protocol": "udp",
			"data": map[string]any{
				"address": pubIP,
				"port":    pubPort,
				"mode":    "xsalsa20_poly1305",
			},
		},
	}

	if err := c.ws.WriteJSON(selectProtocol); err != nil {
		return fmt.Errorf("failed to send select protocol payload: %w", err)
	}

	// Read Opcode 4 (Session Description)
	var sessionDesc struct {
		Op int `json:"op"`
		D  struct {
			Mode      string   `json:"mode"`
			SecretKey []byte   `json:"secret_key"`
		} `json:"d"`
	}

	if err := c.ws.ReadJSON(&sessionDesc); err != nil {
		return fmt.Errorf("failed to read session description: %w", err)
	}

	copy(c.secretKey[:], sessionDesc.D.SecretKey)
	c.udp.SetSecretKey(c.secretKey)

	log.Printf("[Voice] Connection established successfully for Guild %s", c.state.GuildID)
	return nil
}

func (c *Conn) heartbeatLoop() {
	ticker := time.NewTicker(c.heartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.mu.Lock()
			if c.ws != nil {
				payload := map[string]any{
					"op": 3,
					"d":  time.Now().UnixNano() / 1e6,
				}
				_ = c.ws.WriteJSON(payload)
			}
			c.mu.Unlock()
		}
	}
}

// SetSpeaking sends Opcode 5 (Speaking) to Discord.
func (c *Conn) SetSpeaking(speaking bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ws == nil {
		return fmt.Errorf("websocket connection is nil")
	}

	flags := 0
	if speaking {
		flags = 1 << 0
	}

	payload := map[string]any{
		"op": 5,
		"d": map[string]any{
			"speaking": flags,
			"delay":    0,
			"ssrc":     c.ssrc,
		},
	}

	return c.ws.WriteJSON(payload)
}

// Close disconnects the WebSocket and UDP connection.
func (c *Conn) Close() {
	close(c.stopChan)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.udp != nil {
		_ = c.udp.Close()
	}
	if c.ws != nil {
		_ = c.ws.Close()
	}
}
