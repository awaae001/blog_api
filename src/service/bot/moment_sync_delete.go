package bot

import (
	"blog_api/src/model"
	momentRepositories "blog_api/src/repositories/moment"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

// DeleteMomentWithSync attempts remote deletion before deleting the local moment.
// A remote API failure is logged but does not block local deletion.
func DeleteMomentWithSync(db *gorm.DB, id int) error {
	moment, err := momentRepositories.GetMomentByID(db, id)
	if err != nil {
		return err
	}

	platform := momentPlatform(moment)
	switch platform {
	case "discord":
		if discordSession != nil && discordSyncDelete {
			err = discordSession.ChannelMessageDelete(
				strconv.FormatInt(moment.ChannelID, 10),
				strconv.FormatInt(moment.MessageID, 10),
			)
		}
	case "telegram":
		if telegramSession != nil && telegramSyncDelete {
			messageID := int(moment.MessageID)
			if int64(messageID) != moment.MessageID {
				err = fmt.Errorf("message ID out of int range: %d", moment.MessageID)
			} else {
				_, err = telegramSession.Request(tgbotapi.DeleteMessageConfig{
					ChatID:    moment.ChannelID,
					MessageID: messageID,
				})
			}
		}
	}
	if err != nil {
		log.Printf("[moment][WARN] %s sync delete failed for moment %d: %v", platform, id, err)
	}
	return db.Delete(&model.Moment{}, id).Error
}

// Local moments have no remote IDs; older bot records may have no message link.
func momentPlatform(moment *model.Moment) string {
	if moment.ChannelID == 0 || moment.MessageID == 0 {
		return ""
	}
	link := strings.ToLower(moment.MessageLink)
	discordLink := strings.Contains(link, "discord.com")
	telegramLink := strings.Contains(link, "t.me")
	switch {
	case discordLink && telegramLink:
		return ""
	case discordLink:
		return "discord"
	case telegramLink:
		return "telegram"
	case moment.GuildID != 0:
		return "discord"
	case moment.ChannelID < 0 || link == "":
		return "telegram"
	default:
		return ""
	}
}
