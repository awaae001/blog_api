package bot

import (
	"blog_api/src/config"
	"blog_api/src/model"
	"blog_api/src/repositories"
	coreService "blog_api/src/service"
	"blog_api/src/service/oss"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

type discordListener struct {
	db            *gorm.DB
	session       *discordgo.Session
	channelID     string
	filterUserIDs map[string]bool
	ossService    oss.OSSService
}

// StartDiscordListener starts the Discord listener in background.
func StartDiscordListener(db *gorm.DB, cfg *model.Config) {
	dCfg := cfg.MomentsIntegrated.Integrated.Discord
	if !cfg.MomentsIntegrated.Enable || !dCfg.Enable || dCfg.BotToken == "" {
		return
	}

	session, err := discordgo.New("Bot " + dCfg.BotToken)
	if err != nil {
		log.Printf("[discord] init session failed: %v", err)
		return
	}
	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	listener := &discordListener{
		db:            db,
		session:       session,
		channelID:     strings.TrimSpace(dCfg.ChannelID),
		filterUserIDs: make(map[string]bool),
	}

	for _, id := range dCfg.FilterUserid {
		listener.filterUserIDs[strconv.FormatInt(id, 10)] = true
	}
	if len(listener.filterUserIDs) == 0 {
		log.Printf("[discord] filter_userid is empty; all users will be accepted")
	}

	if cfg.OSS.Enable {
		if ossService, err := oss.NewOSSService(); err == nil {
			listener.ossService = ossService
		} else {
			log.Printf("[discord] oss init failed: %v", err)
		}
	}

	session.AddHandler(listener.onMessageCreate)
	if err := session.Open(); err != nil {
		log.Printf("[discord] open session failed: %v", err)
		return
	}

	log.Println("[discord] listener started")
}

func (l *discordListener) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m == nil || m.Message == nil || m.Author == nil {
		return
	}

	if s.State != nil && s.State.User != nil && m.Author.ID == s.State.User.ID {
		return
	}

	if l.channelID != "" && m.ChannelID != l.channelID {
		return
	}

	if len(l.filterUserIDs) > 0 && !l.filterUserIDs[m.Author.ID] {
		log.Printf("[discord] filtered message from %s", m.Author.ID)
		return
	}

	channelID, err := parseDiscordID(m.ChannelID)
	if err != nil {
		log.Printf("[discord] invalid channel id: %v", err)
		return
	}

	messageID, err := parseDiscordID(m.ID)
	if err != nil {
		log.Printf("[discord] invalid message id: %v", err)
		return
	}

	var guildID int64
	if m.GuildID != "" {
		if parsed, err := parseDiscordID(m.GuildID); err == nil {
			guildID = parsed
		}
	}

	content := m.Content
	media := l.downloadAttachments(m.Attachments)
	l.saveMoment(guildID, channelID, messageID, m.Timestamp.Unix(), content, media)
}

func (l *discordListener) downloadAttachments(attachments []*discordgo.MessageAttachment) []model.MomentMedia {
	if len(attachments) == 0 {
		return nil
	}

	var media []model.MomentMedia
	for _, att := range attachments {
		mediaType := detectDiscordMediaType(att)
		if mediaType == "" {
			continue
		}

		item, err := l.downloadAttachment(att, mediaType)
		if err != nil {
			log.Printf("[discord] download attachment failed: %v", err)
			continue
		}
		if item != nil {
			media = append(media, *item)
		}
	}

	return media
}

func detectDiscordMediaType(att *discordgo.MessageAttachment) string {
	if att == nil {
		return ""
	}
	contentType := strings.TrimSpace(att.ContentType)
	if strings.HasPrefix(contentType, "image/") {
		return "image"
	}
	if strings.HasPrefix(contentType, "video/") {
		return "video"
	}

	ext := strings.ToLower(filepath.Ext(att.Filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
		return "image"
	case ".mp4", ".webm":
		return "video"
	default:
		return ""
	}
}

func (l *discordListener) downloadAttachment(att *discordgo.MessageAttachment, mediaType string) (*model.MomentMedia, error) {
	if att == nil || att.URL == "" {
		return nil, nil
	}

	resp, err := http.Get(att.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return nil, fmt.Errorf("discord download failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fileName, contentType, err := normalizeDiscordFile(att.Filename, att.ContentType, mediaType, resp, data)
	if err != nil {
		return nil, err
	}

	storedURL, err := l.storeFile(fileName, contentType, data)
	if err != nil {
		return nil, err
	}

	return &model.MomentMedia{
		Name:      fileName,
		MediaURL:  storedURL,
		MediaType: mediaType,
	}, nil
}

func normalizeDiscordFile(fileName, mimeType, mediaType string, resp *http.Response, data []byte) (string, string, error) {
	contentType := strings.TrimSpace(mimeType)
	if contentType == "" && resp != nil {
		contentType = strings.TrimSpace(resp.Header.Get("Content-Type"))
	}
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	detectedType := http.DetectContentType(data)
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = detectedType
	}

	switch mediaType {
	case "image":
		if !strings.HasPrefix(contentType, "image/") && !strings.HasPrefix(detectedType, "image/") {
			return "", "", fmt.Errorf("unexpected content type for image: %s", contentType)
		}
	case "video":
		if !strings.HasPrefix(contentType, "video/") && !strings.HasPrefix(detectedType, "video/") {
			return "", "", fmt.Errorf("unexpected content type for video: %s", contentType)
		}
	}

	if fileName == "" {
		fileName = "discord"
	}
	if filepath.Ext(fileName) == "" {
		exts, _ := mime.ExtensionsByType(contentType)
		if len(exts) == 0 {
			exts, _ = mime.ExtensionsByType(detectedType)
		}
		if len(exts) > 0 {
			fileName += exts[0]
		}
	}

	return fileName, contentType, nil
}

func (l *discordListener) storeFile(name, mimeType string, data []byte) (string, error) {
	datePath := time.Now().Format("060102")
	finalSubPath := filepath.Join("dis", datePath)
	if l.ossService != nil {
		path := filepath.Join(finalSubPath, name)
		return uploadToOSS(l.ossService, path, mimeType, data)
	}

	svc := coreService.NewResourceService(config.GetConfig())
	_, url, err := svc.SaveBytes(name, data, finalSubPath, false)
	return url, err
}

func (l *discordListener) saveMoment(guildID, channelID, msgID, date int64, content string, media []model.MomentMedia) {
	if content == "" && len(media) == 0 {
		return
	}

	exists, err := repositories.MomentExistsByChannelMessage(l.db, channelID, msgID)
	if err != nil || exists {
		return
	}

	moment := model.Moment{
		Content:   content,
		Status:    "visible",
		GuildID:   guildID,
		ChannelID: channelID,
		MessageID: msgID,
		CreatedAt: date,
	}

	if err := repositories.CreateMoment(l.db, &moment, media); err != nil {
		log.Printf("[discord] create moment failed: %v", err)
	} else {
		log.Printf("[discord] saved moment channel=%d msg=%d media=%d", channelID, msgID, len(media))
	}
}

func parseDiscordID(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("empty id")
	}
	return strconv.ParseInt(raw, 10, 64)
}
