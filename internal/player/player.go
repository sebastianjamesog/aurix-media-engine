package player

import (
	"sync"
	"time"
)

// Track represents a resolved audio track item.
type Track struct {
	Encoded    string `json:"encoded"`
	Identifier string `json:"identifier"`
	Title      string `json:"title"`
	Author     string `json:"author"`
	URI        string `json:"uri"`
	Length     int64  `json:"length"` // in milliseconds
	IsStream   bool   `json:"isStream"`
	Position   int64  `json:"position"`
}

// Player state for a single guild.
type Player struct {
	GuildID   string `json:"guildId"`
	Volume    int    `json:"volume"`
	Paused    bool   `json:"paused"`
	Track     *Track `json:"track,omitempty"`
	Position  int64  `json:"position"`
	Connected bool   `json:"connected"`
	mu        sync.RWMutex
}

// Manager maintains active players across guilds.
type Manager struct {
	players sync.Map // map[string]*Player
}

// NewManager creates a new player manager instance.
func NewManager() *Manager {
	return &Manager{}
}

// GetOrCreate returns an existing player or creates a new one for the guild.
func (m *Manager) GetOrCreate(guildID string) *Player {
	if p, ok := m.players.Load(guildID); ok {
		return p.(*Player)
	}

	p := &Player{
		GuildID:   guildID,
		Volume:    100,
		Paused:    false,
		Connected: false,
	}

	actual, _ := m.players.LoadOrStore(guildID, p)
	return actual.(*Player)
}

// Destroy removes a player for a guild.
func (m *Manager) Destroy(guildID string) bool {
	_, loaded := m.players.LoadAndDelete(guildID)
	return loaded
}

// List returns a snapshot of all active players.
func (m *Manager) List() []*Player {
	var list []*Player
	m.players.Range(func(key, value any) bool {
		list = append(list, value.(*Player))
		return true
	})
	return list
}

// Stats holds telemetry information about the engine.
type Stats struct {
	Players        int   `json:"players"`
	PlayingPlayers int   `json:"playingPlayers"`
	Uptime         int64 `json:"uptime"`
}

var startTime = time.Now()

// GetStats returns engine runtime statistics.
func (m *Manager) GetStats() Stats {
	total := 0
	playing := 0
	m.players.Range(func(key, value any) bool {
		total++
		p := value.(*Player)
		p.mu.RLock()
		if p.Track != nil && !p.Paused {
			playing++
		}
		p.mu.RUnlock()
		return true
	})

	return Stats{
		Players:        total,
		PlayingPlayers: playing,
		Uptime:         time.Since(startTime).Milliseconds(),
	}
}
