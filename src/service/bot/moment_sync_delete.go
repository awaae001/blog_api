package bot

import (
	"blog_api/src/config"
	"blog_api/src/model"
	momentRepositories "blog_api/src/repositories/moment"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

// DeleteMomentWithSync deletes a moment locally and attempts to delete it from social platforms.
// Remote delete errors are logged but do not block local deletion.
func DeleteMomentWithSync(db *gorm.DB, id int) error {
	moment, err := momentRepositories.GetMomentByID(db, id)
	if err != nil {
		return err
	}

	if err := syncDeleteMoment(config.GetConfig(), moment); err != nil {
		log.Printf("[moment][WARN] sync delete failed for moment %d: %v", id, err)
	}

	return momentRepositories.DeleteMoment(db, id)
}

func syncDeleteMoment(cfg *model.Config, moment *model.Moment) error {
	if cfg == nil || moment == nil {
		return nil
	}

	var errs []string
	if shouldDeleteDiscord(cfg, moment) {
		if err := deleteDiscordMessage(cfg, moment); err != nil {
			errs = append(errs, fmt.Sprintf("discord: %v", err))
		}
	}
	if shouldDeleteTelegram(cfg, moment) {
		if err := deleteTelegramMessage(cfg, moment); err != nil {
			errs = append(errs, fmt.Sprintf("telegram: %v", err))
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func shouldDeleteDiscord(cfg *model.Config, moment *model.Moment) bool {
	if cfg == nil || moment == nil {
		return false
	}
	dCfg := cfg.MomentsIntegrated.Integrated.Discord
	if !cfg.MomentsIntegrated.Enable || !dCfg.Enable || !dCfg.SyncDelete || dCfg.BotToken == "" {
		return false
	}
	if moment.ChannelID == 0 || moment.MessageID == 0 {
		return false
	}
	link := strings.ToLower(moment.MessageLink)
	if strings.Contains(link, "t.me") {
		return false
	}
	return strings.Contains(link, "discord.com") || moment.GuildID != 0
}

func shouldDeleteTelegram(cfg *model.Config, moment *model.Moment) bool {
	if cfg == nil || moment == nil {
		return false
	}
	tgCfg := cfg.MomentsIntegrated.Integrated.Telegram
	if !cfg.MomentsIntegrated.Enable || !tgCfg.Enable || !tgCfg.SyncDelete || tgCfg.BotToken == "" {
		return false
	}
	if moment.ChannelID == 0 || moment.MessageID == 0 {
		return false
	}
	link := strings.ToLower(moment.MessageLink)
	if strings.Contains(link, "discord.com") {
		return false
	}
	if strings.Contains(link, "t.me") {
		return true
	}
	if moment.ChannelID < 0 {
		return true
	}
	return moment.GuildID == 0 && link == ""
}

func deleteDiscordMessage(cfg *model.Config, moment *model.Moment) error {
	channelID := strconv.FormatInt(moment.ChannelID, 10)
	messageID := strconv.FormatInt(moment.MessageID, 10)
	session := GetDiscordSession()
	if session == nil {
		return fmt.Errorf("discord session not initialized")
	}
	return session.ChannelMessageDelete(channelID, messageID)
}

func deleteTelegramMessage(cfg *model.Config, moment *model.Moment) error {
	tgBot := GetTelegramBot()
	if tgBot == nil {
		return fmt.Errorf("telegram bot not initialized")
	}
	messageID, err := safeIntFromInt64(moment.MessageID)
	if err != nil {
		return err
	}
	_, err = tgBot.Request(tgbotapi.DeleteMessageConfig{
		ChatID:    moment.ChannelID,
		MessageID: messageID,
	})
	return err
}

func safeIntFromInt64(val int64) (int, error) {
	maxInt := int64(^uint(0) >> 1)
	if val > maxInt || val < -maxInt-1 {
		return 0, fmt.Errorf("value out of int range: %d", val)
	}
	return int(val), nil
}
