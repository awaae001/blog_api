package bot

import (
	"blog_api/src/model"

	"gorm.io/gorm"
)

// StartListeners starts all bot listeners with a shared config.
func StartListeners(db *gorm.DB, cfg *model.Config) {
	if cfg == nil {
		return
	}
	StartTelegramListener(db, cfg)
	StartDiscordListener(db, cfg)
}
