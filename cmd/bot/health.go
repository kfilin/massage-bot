package main

import (
	"github.com/kfilin/massage-bot/internal/logging"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "healthy", "service": "massage-bot"}`)
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "ready", "service": "massage-bot"}`)
}

func liveHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "live", "service": "massage-bot"}`)
}

func startHealthServer(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/ready", readyHandler)
	mux.HandleFunc("/live", liveHandler)
	mux.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"service": "massage-bot", "endpoints": ["/health", "/ready", "/live"]}`)
	})

	// Get port from environment or use default
	port := os.Getenv("HEALTH_PORT")
	if port == "" {
		port = "8083"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	logging.Infof("Starting health server on :%s", port)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Health server failed to start: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down Health server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logging.Infof("Health server shutdown error: %v", err)
	}
}
