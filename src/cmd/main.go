package cmd

import (
	"blog_api/src/config"
	"blog_api/src/handler"
	"blog_api/src/repositories"
	"fmt"
	"log"
)

// Run starts the application
func Run() {
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

	// Setup HTTP router
	router := handler.SetupRouter(db, cfg)

	// Start HTTP server in a separate goroutine
	go func() {
		addr := fmt.Sprintf("%s:%s", cfg.ListenAddress, cfg.Port)
		log.Printf("Starting HTTP server on %s", addr)
		if err := router.Run(addr); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Start the cron jobs
	StartCronJobs(db)
	log.Println("Application started successfully. HTTP server and cron jobs are running.")

	// Block forever
	select {}
}
