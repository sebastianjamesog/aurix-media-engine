package gateway

import (
	"encoding/json"
	"net/http"

	"aurix/internal/dashboard"
	"aurix/internal/player"
	"aurix/internal/providers"
)

// Server handles REST endpoints and API Gateway operations.
type Server struct {
	pm        *player.Manager
	registry  *providers.Registry
	password  string
}

// NewServer creates a new Server instance.
func NewServer(pm *player.Manager, registry *providers.Registry, password string) *Server {
	return &Server{
		pm:       pm,
		registry: registry,
		password: password,
	}
}

// RegisterRoutes attaches REST and WebSocket endpoints to an HTTP Mux.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	dashboard.RegisterRoutes(mux)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") == "websocket" {
			s.HandleWebSocket(w, r)
			return
		}
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	})
	mux.HandleFunc("GET /v4/websocket", s.HandleWebSocket)
	mux.HandleFunc("GET /v4/stats", s.authMiddleware(s.handleStats))
	mux.HandleFunc("GET /v4/loadtracks", s.authMiddleware(s.handleLoadTracks))
	mux.HandleFunc("GET /v4/sessions/{sessionId}/players", s.authMiddleware(s.handleGetPlayers))
	mux.HandleFunc("DELETE /v4/sessions/{sessionId}/players/{guildId}", s.authMiddleware(s.handleDestroyPlayer))
}

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != s.password {
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats := s.pm.GetStats()
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleLoadTracks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	identifier := r.URL.Query().Get("identifier")

	if identifier == "" {
		http.Error(w, `{"loadType":"empty","data":{}}`, http.StatusBadRequest)
		return
	}

	if s.registry != nil {
		res, err := s.registry.Resolve(r.Context(), identifier)
		if err != nil {
			http.Error(w, `{"loadType":"error","data":{}}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	resp := map[string]any{
		"loadType": "empty",
		"data":     map[string]any{},
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleGetPlayers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	players := s.pm.List()
	json.NewEncoder(w).Encode(players)
}

func (s *Server) handleDestroyPlayer(w http.ResponseWriter, r *http.Request) {
	guildID := r.PathValue("guildId")
	if guildID == "" {
		http.Error(w, `{"error":"Missing guildId"}`, http.StatusBadRequest)
		return
	}

	s.pm.Destroy(guildID)
	w.WriteHeader(http.StatusNoContent)
}
