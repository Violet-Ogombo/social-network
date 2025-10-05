package main

import (
	"net/http"
	"os"
	"social-network/backend/handlers"
)

func registerRoutes(mux *http.ServeMux) {
	// Serve production build if present, otherwise the dev public folder
	if _, err := os.Stat("./frontend/dist"); err == nil {
		mux.Handle("/", http.FileServer(http.Dir("./frontend/dist")))
	} else {
		mux.Handle("/", http.FileServer(http.Dir("./frontend/public")))
	}

	// Websocket endpoint (protected by auth middleware so context contains user ID)
	mux.Handle("/ws", AuthMiddleware(http.HandlerFunc(HandleWebSocket)))

	// API endpoints
	mux.HandleFunc("/register", handlers.RegisterHandler)
	mux.HandleFunc("/login", handlers.LoginHandler)
	mux.HandleFunc("/logout", handlers.LogoutHandler)
	mux.HandleFunc("/api/check-session", handlers.CheckSessionHandler)
}
