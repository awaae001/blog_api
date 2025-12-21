package main

import (
	"blog_api/src/config"
	"blog_api/src/repositories"
	botService "blog_api/src/service/bot"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[bot] failed to load configuration: %v", err)
	}

	db, err := repositories.InitDB(cfg)
	if err != nil {
		log.Fatalf("[bot] failed to initialize database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("[bot] failed to get sql.DB from gorm: %v", err)
	}
	defer sqlDB.Close()

	if !cfg.MomentsIntegrated.Enable || !cfg.MomentsIntegrated.Integrated.Telegram.Enable {
		log.Println("[bot] integration disabled in config, exiting")
		return
	}

	botService.StartTelegramListener(db)

	log.Println("[bot] listener started")
	select {}
}
