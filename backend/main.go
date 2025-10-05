package main

import (
	"log"
	"net/http"
	"time"

	"social-network/backend/db"
	"social-network/backend/handlers"
	"social-network/backend/utils"
)

func main() {
	db.InitDB() // connect + run migrations
	// inject DB into utils package for session helpers
	utils.SetDB(db.DB)

	mux := http.NewServeMux()
	registerRoutes(mux)

	// Start periodic session cleanup
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			handlers.CleanupSessions()
		}
	}()

	log.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
