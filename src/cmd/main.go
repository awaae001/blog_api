package main

import (
	"blog_api/src/config"
	"blog_api/src/repositories"
	"log"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := repositories.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Insert friend links from config
	if err := repositories.InsertFriendLinks(db, cfg.FriendLinks); err != nil {
		log.Printf("Could not insert friend links: %v", err)
	}

	log.Println("Application started successfully.")
}
