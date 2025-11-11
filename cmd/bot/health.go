package main

import (
    "fmt"
    "log"
    "net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, `{"status": "healthy", "service": "massage-bot"}`)
}

func startHealthServer() {
    http.HandleFunc("/health", healthHandler)
    
    log.Println("Starting health server on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Health server failed to start: %v", err)
    }
}
