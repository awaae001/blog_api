package cmd

import (
	cmd "blog_api/src/cmd/router"
	"blog_api/src/config"
	"blog_api/src/repositories"
	friendsRepositories "blog_api/src/repositories/friend"
	"blog_api/src/service"
	botService "blog_api/src/service/bot"
	"blog_api/src/service/oss"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

// Run 启动应用程序
func Run() {
	startTime := time.Now()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[main]加载配置失败: %v", err)
	}
	db, err := repositories.InitDB(cfg)
	if err != nil {
		log.Fatalf("[main]初始化数据库失败: %v", err)
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

	addr := fmt.Sprintf("%s:%s", cfg.ListenAddress, cfg.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       2 * time.Minute,
		WriteTimeout:      2 * time.Minute,
		IdleTimeout:       time.Minute,
	}
	go func() {
		log.Printf("[main][Http]HTTP 服务器启动于 %s", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[main][Http]启动 HTTP 服务器失败: %v", err)
		}
	}()

	botService.StartListeners(db, cfg)
	StartCronJobs(db)
	log.Println("[main][App]应用程序启动成功。HTTP 服务器和 cron 任务正在运行。")

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-sigCtx.Done()

	log.Println("[main][App]收到退出信号，开始优雅关闭...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[main][Http]优雅关闭失败，强制退出: %v", err)
	}
	log.Println("[main][App]HTTP 服务器已关闭，进程退出。")
}
