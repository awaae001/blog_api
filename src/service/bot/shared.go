package bot

import (
	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Initialized before the HTTP server starts; configuration changes require a restart.
var (
	discordSession     *discordgo.Session
	telegramSession    *tgbotapi.BotAPI
	discordSyncDelete  bool
	telegramSyncDelete bool
)
