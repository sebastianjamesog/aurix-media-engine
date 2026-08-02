package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aurix/internal/gateway"
	"aurix/internal/player"
	"aurix/internal/providers"
)

const version = "0.1.0-alpha"

func main() {
	log.Printf("Starting Aurix Media Engine (AME) v%s...", version)

	// Initialize Player Manager
	pm := player.NewManager()

	// Initialize Provider Registry
	registry := providers.NewRegistry()
	registry.Register(providers.NewHTTPProvider())
	registry.Register(providers.NewLocalProvider("."))
	registry.Register(providers.NewYouTubeProvider())
	registry.Register(providers.NewSpotifyProvider())

	// Initialize API Gateway
	gw := gateway.NewServer(pm, registry, "youshallnotpass")

	mux := http.NewServeMux()
	gw.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         ":2333",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Channel to listen for shutdown signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("AME HTTP & WebSocket Gateway listening on http://localhost:2333")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stopChan
	log.Println("Shutting down Aurix Media Engine gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Aurix Media Engine stopped cleanly.")
}
