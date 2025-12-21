package bot

import (
	"blog_api/src/config"
	"blog_api/src/model"
	"blog_api/src/repositories"
	coreService "blog_api/src/service"
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

const defaultTelegramMediaSubPath = "telegram"

type telegramMediaRef struct {
	FileID    string
	FileName  string
	MimeType  string
	MediaType string
}

type telegramMediaGroup struct {
	Messages []*tgbotapi.Message
	LastSeen time.Time
}

type telegramListener struct {
	db              *gorm.DB
	bot             *tgbotapi.BotAPI
	channelID       int64
	channelUsername string
	filterUserIDs   []int64
	ossService      coreService.OSSService
	useOSS          bool
	mediaSubPath    string
	pendingGroups   map[string]*telegramMediaGroup
}

// StartTelegramListener starts the Telegram listener in background.
func StartTelegramListener(db *gorm.DB) {
	cfg := config.GetConfig()
	if !cfg.MomentsIntegrated.Enable || !cfg.MomentsIntegrated.Integrated.Telegram.Enable {
		return
	}

	token := strings.TrimSpace(cfg.MomentsIntegrated.Integrated.Telegram.BotToken)
	if token == "" {
		log.Println("[telegram] bot token is empty, skipping integration")
		return
	}

	mediaSubPath := strings.TrimSpace(cfg.MomentsIntegrated.Integrated.Telegram.MediaPath)
	if mediaSubPath == "" {
		mediaSubPath = defaultTelegramMediaSubPath
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Printf("[telegram] init bot failed: %v", err)
		return
	}

	listener := &telegramListener{
		db:            db,
		bot:           bot,
		filterUserIDs: cfg.MomentsIntegrated.Integrated.Telegram.FilterUserid,
		mediaSubPath:  mediaSubPath,
		pendingGroups: make(map[string]*telegramMediaGroup),
	}

	channelID, channelUsername, err := parseTelegramChannel(cfg.MomentsIntegrated.Integrated.Telegram.ChannelID)
	if err != nil {
		log.Printf("[telegram] invalid channel id: %v", err)
		return
	}
	listener.channelID = channelID
	listener.channelUsername = channelUsername

	if cfg.OSS.Enable {
		if err := coreService.ValidateOSSConfig(); err != nil {
			log.Printf("[telegram] oss validation failed, falling back to local: %v", err)
		} else {
			ossService, err := coreService.NewOSSService()
			if err != nil {
				log.Printf("[telegram] oss init failed, falling back to local: %v", err)
			} else {
				listener.ossService = ossService
				listener.useOSS = true
			}
		}
	}

	go listener.run()
}

func parseTelegramChannel(channelID string) (int64, string, error) {
	raw := strings.TrimSpace(channelID)
	if raw == "" {
		return 0, "", nil
	}
	if strings.HasPrefix(raw, "@") {
		return 0, strings.TrimPrefix(raw, "@"), nil
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, "", err
	}
	return id, "", nil
}

func (l *telegramListener) run() {
	log.Println("[telegram] listener started")
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := l.bot.GetUpdatesChan(updateConfig)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case update, ok := <-updates:
			if !ok {
				log.Println("[telegram] updates channel closed")
				return
			}
			msg := pickMessage(update)
			if msg == nil {
				continue
			}
			if !l.matchChannel(msg.Chat) || !l.matchUser(msg) {
				continue
			}
			if msg.MediaGroupID != "" {
				l.collectGroup(msg, time.Now())
				continue
			}
			l.handleSingleMessage(msg)
		case <-ticker.C:
			l.flushGroups(2 * time.Second)
		}
	}
}

func pickMessage(update tgbotapi.Update) *tgbotapi.Message {
	if update.Message != nil {
		return update.Message
	}
	if update.ChannelPost != nil {
		return update.ChannelPost
	}
	return nil
}

func (l *telegramListener) collectGroup(msg *tgbotapi.Message, now time.Time) {
	group := l.pendingGroups[msg.MediaGroupID]
	if group == nil {
		group = &telegramMediaGroup{}
		l.pendingGroups[msg.MediaGroupID] = group
	}
	group.Messages = append(group.Messages, msg)
	group.LastSeen = now
}

func (l *telegramListener) flushGroups(idle time.Duration) {
	if len(l.pendingGroups) == 0 {
		return
	}
	now := time.Now()
	for key, group := range l.pendingGroups {
		if now.Sub(group.LastSeen) < idle {
			continue
		}
		l.handleMediaGroup(group.Messages)
		delete(l.pendingGroups, key)
	}
}

func (l *telegramListener) handleSingleMessage(msg *tgbotapi.Message) {
	log.Printf("[telegram] received message chat=%d message=%d user=%s", msg.Chat.ID, msg.MessageID, describeSender(msg))
	content := resolveMessageContent(msg)
	media, err := l.buildMedia(msg)
	if err != nil {
		log.Printf("[telegram] media handling failed: %v", err)
		return
	}

	if content == "" && len(media) == 0 {
		return
	}

	exists, err := repositories.MomentExistsByChannelMessage(l.db, msg.Chat.ID, int64(msg.MessageID))
	if err != nil {
		log.Printf("[telegram] check existing moment failed: %v", err)
		return
	}
	if exists {
		return
	}

	moment := model.Moment{
		Content:   content,
		Status:    "visible",
		ChannelID: msg.Chat.ID,
		MessageID: int64(msg.MessageID),
		CreatedAt: int64(msg.Date),
	}

	if err := repositories.CreateMoment(l.db, &moment, media); err != nil {
		log.Printf("[telegram] create moment failed: %v", err)
	}
}

func (l *telegramListener) handleMediaGroup(messages []*tgbotapi.Message) {
	if len(messages) == 0 {
		return
	}
	content := ""
	media := make([]model.MomentMedia, 0, len(messages))
	minMessageID := int64(messages[0].MessageID)
	minDate := messages[0].Date
	chatID := messages[0].Chat.ID
	log.Printf("[telegram] received media group chat=%d messages=%d", chatID, len(messages))

	for _, msg := range messages {
		if int64(msg.MessageID) < minMessageID {
			minMessageID = int64(msg.MessageID)
		}
		if msg.Date < minDate {
			minDate = msg.Date
		}
		if content == "" {
			content = resolveMessageContent(msg)
		}
		if msg.Chat.ID != chatID {
			continue
		}
		msgMedia, err := l.buildMedia(msg)
		if err != nil {
			log.Printf("[telegram] media group item failed: %v", err)
			return
		}
		media = append(media, msgMedia...)
	}

	if content == "" && len(media) == 0 {
		return
	}

	exists, err := repositories.MomentExistsByChannelMessage(l.db, chatID, minMessageID)
	if err != nil {
		log.Printf("[telegram] check existing moment failed: %v", err)
		return
	}
	if exists {
		return
	}

	moment := model.Moment{
		Content:   content,
		Status:    "visible",
		ChannelID: chatID,
		MessageID: minMessageID,
		CreatedAt: int64(minDate),
	}

	if err := repositories.CreateMoment(l.db, &moment, media); err != nil {
		log.Printf("[telegram] create moment failed: %v", err)
	}
}

func resolveMessageContent(msg *tgbotapi.Message) string {
	content := strings.TrimSpace(msg.Text)
	if content == "" {
		content = strings.TrimSpace(msg.Caption)
	}
	return content
}

func (l *telegramListener) matchChannel(chat *tgbotapi.Chat) bool {
	if l.channelID != 0 {
		return chat.ID == l.channelID
	}
	if l.channelUsername != "" {
		return strings.EqualFold(chat.UserName, l.channelUsername)
	}
	return true
}

func (l *telegramListener) matchUser(msg *tgbotapi.Message) bool {
	// Channel posts should always be accepted.
	if msg.SenderChat != nil && msg.SenderChat.Type == "channel" {
		return true
	}

	if len(l.filterUserIDs) == 0 {
		return true
	}

	if msg.From != nil {
		for _, id := range l.filterUserIDs {
			if id == msg.From.ID {
				return true
			}
		}
		log.Printf("[telegram] message filtered by user list chat=%d message=%d user=%d", msg.Chat.ID, msg.MessageID, msg.From.ID)
		return false
	}

	if msg.SenderChat != nil {
		for _, id := range l.filterUserIDs {
			if id == msg.SenderChat.ID {
				return true
			}
		}
		log.Printf("[telegram] message filtered by sender chat list chat=%d message=%d sender=%d", msg.Chat.ID, msg.MessageID, msg.SenderChat.ID)
	}

	return false
}

func describeSender(msg *tgbotapi.Message) string {
	if msg.From != nil {
		return fmt.Sprintf("user:%d", msg.From.ID)
	}
	if msg.SenderChat != nil {
		return fmt.Sprintf("chat:%d", msg.SenderChat.ID)
	}
	return "unknown"
}

func (l *telegramListener) buildMedia(msg *tgbotapi.Message) ([]model.MomentMedia, error) {
	ref, ok := pickMediaRef(msg)
	if !ok {
		return nil, nil
	}

	fileName, data, mimeType, err := l.downloadFile(ref)
	if err != nil {
		return nil, err
	}

	mediaURL, err := l.storeMedia(fileName, mimeType, data)
	if err != nil {
		return nil, err
	}

	return []model.MomentMedia{
		{
			Name:      filepath.Base(fileName),
			MediaURL:  mediaURL,
			MediaType: ref.MediaType,
			IsDeleted: 0,
		},
	}, nil
}

func pickMediaRef(msg *tgbotapi.Message) (telegramMediaRef, bool) {
	if len(msg.Photo) > 0 {
		best := msg.Photo[0]
		for _, p := range msg.Photo {
			if p.FileSize > best.FileSize {
				best = p
			}
		}
		return telegramMediaRef{
			FileID:    best.FileID,
			MimeType:  "image/jpeg",
			MediaType: "image",
		}, true
	}

	if msg.Video != nil {
		return telegramMediaRef{
			FileID:    msg.Video.FileID,
			FileName:  msg.Video.FileName,
			MimeType:  msg.Video.MimeType,
			MediaType: "video",
		}, true
	}

	if msg.Animation != nil {
		return telegramMediaRef{
			FileID:    msg.Animation.FileID,
			FileName:  msg.Animation.FileName,
			MimeType:  msg.Animation.MimeType,
			MediaType: "video",
		}, true
	}

	if msg.Document != nil {
		mediaType := "image"
		if strings.HasPrefix(msg.Document.MimeType, "video/") {
			mediaType = "video"
		}
		return telegramMediaRef{
			FileID:    msg.Document.FileID,
			FileName:  msg.Document.FileName,
			MimeType:  msg.Document.MimeType,
			MediaType: mediaType,
		}, true
	}

	return telegramMediaRef{}, false
}

func (l *telegramListener) downloadFile(ref telegramMediaRef) (string, []byte, string, error) {
	fileName := filepath.Base(ref.FileName)
	mimeType := ref.MimeType

	file, err := l.bot.GetFile(tgbotapi.FileConfig{FileID: ref.FileID})
	if err != nil {
		return "", nil, "", err
	}

	if fileName == "" {
		fileName = filepath.Base(file.FilePath)
	}
	if mimeType == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(fileName))
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if filepath.Ext(fileName) == "" {
		extensions, err := mime.ExtensionsByType(mimeType)
		if err == nil && len(extensions) > 0 {
			fileName = fileName + extensions[0]
		}
	}

	url := file.Link(l.bot.Token)
	resp, err := http.Get(url)
	if err != nil {
		return "", nil, "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, "", err
	}
	if len(data) == 0 {
		return "", nil, "", fmt.Errorf("telegram file is empty")
	}
	return fileName, data, mimeType, nil
}

func (l *telegramListener) storeMedia(fileName, mimeType string, data []byte) (string, error) {
	safeName := filepath.Base(fileName)
	if safeName == "" {
		return "", fmt.Errorf("empty file name")
	}

	if l.useOSS && l.ossService != nil {
		ossFileName := safeName
		if l.mediaSubPath != "" {
			ossFileName = path.Join(strings.TrimPrefix(l.mediaSubPath, "/"), safeName)
		}
		return uploadWithOSS(l.ossService, ossFileName, mimeType, data)
	}

	resourceService := coreService.NewResourceService(config.GetConfig())
	_, urlPath, err := resourceService.SaveBytes(safeName, data, l.mediaSubPath, false)
	if err != nil {
		return "", err
	}
	return urlPath, nil
}

func uploadWithOSS(service coreService.OSSService, fileName, mimeType string, data []byte) (string, error) {
	reader := bytes.NewReader(data)
	file := &readSeekCloser{Reader: reader}
	header := &multipart.FileHeader{
		Filename: fileName,
		Header:   make(textproto.MIMEHeader),
		Size:     int64(len(data)),
	}
	if mimeType != "" {
		header.Header.Set("Content-Type", mimeType)
	}
	return service.UploadFile(file, header)
}

type readSeekCloser struct {
	*bytes.Reader
}

func (r *readSeekCloser) Close() error {
	return nil
}
