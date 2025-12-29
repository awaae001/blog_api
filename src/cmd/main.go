package cmd

import (
	cmd "blog_api/src/cmd/router"
	"blog_api/src/config"
	"blog_api/src/repositories"
	friendsRepositories "blog_api/src/repositories/friend"
	"blog_api/src/service"
	botService "blog_api/src/service/bot"
	"blog_api/src/service/oss"
	"fmt"
	"log"
	"time"
)

// Run starts the application
func Run() {
	startTime := time.Now()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[main]Failed to load configuration: %v", err)
	}
	db, err := repositories.InitDB(cfg)
	if err != nil {
		log.Fatalf("[main]Failed to initialize database: %v", err)
	}
	if err := friendsRepositories.InsertFriendLinks(db, cfg.FriendLinks); err != nil {
		log.Printf("[main]无法插入友链: %v", err)
	}
	if err := service.ScanAndSaveImages(db); err != nil {
		log.Printf("[main]无法扫描和保存图片: %v", err)
	}
	if err := oss.ValidateOSSConfig(); err != nil {
		log.Printf("[main][OSS]配置校验失败: %v", err)
	}
	router := cmd.SetupRouter(db, cfg, startTime)

	go func() {
		addr := fmt.Sprintf("%s:%s", cfg.ListenAddress, cfg.Port)
		log.Printf("[main][Http]HTTP 服务器启动于 %s", addr)
		if err := router.Run(addr); err != nil {
			log.Fatalf("[main][Http]Failed to start HTTP server: %v", err)
		}
	}()

	botService.StartListeners(db, cfg)
	StartCronJobs(db)
	log.Println("[main][App]Application started successfully. HTTP server and cron jobs are running.")

	select {}
}
