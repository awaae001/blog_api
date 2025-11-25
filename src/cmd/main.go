package cmd

import (
	cmd "blog_api/src/cmd/router"
	"blog_api/src/config"
	"blog_api/src/repositories"
	"fmt"
	"log"
	"time"
)

// Run starts the application
func Run() {
	startTime := time.Now()
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[main]Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := repositories.InitDB(cfg)
	if err != nil {
		log.Fatalf("[main]Failed to initialize database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("[main]Failed to get sql.DB from gorm: %v", err)
	}
	defer sqlDB.Close()

	// Insert friend links from config
	if err := repositories.InsertFriendLinks(db, cfg.FriendLinks); err != nil {
		log.Printf("[main]无法插入友链: %v", err)
	}

	// Setup HTTP router
	router := cmd.SetupRouter(db, cfg, startTime)

	// Start HTTP server in a separate goroutine
	go func() {
		addr := fmt.Sprintf("%s:%s", cfg.ListenAddress, cfg.Port)
		log.Printf("[main][Http]HTTP 服务器启动于 %s", addr)
		if err := router.Run(addr); err != nil {
			log.Fatalf("[main][Http]Failed to start HTTP server: %v", err)
		}
	}()

	// Start the cron jobs
	StartCronJobs(db)
	log.Println("[main][App]Application started successfully. HTTP server and cron jobs are running.")

	// Block forever
	select {}
}
