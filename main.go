package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"web-analyzer/handler"
	"web-analyzer/middleware"
)

func main() {
	mux := http.NewServeMux()

	// Static file for HTML frontend
	mux.Handle("/", http.FileServer(http.Dir("./web")))
	// API Endpoint for URL analysis
	mux.HandleFunc("/analyze", handler.AnalyzeHandler)

	server := &http.Server{
		Addr: ":8080",
		//Handler: mux,
		Handler: middleware.Logging(mux),
	}

	log.Println("Server started on http://localhost:8080")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server failes : %v", err)
	}

	// wait for interrupt signal (Ctrl+C)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
	log.Println("Server gracefully stopped")

}
