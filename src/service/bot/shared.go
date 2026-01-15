package bot

import (
	"blog_api/src/service/oss"
	"bytes"
	"mime/multipart"
	"net/textproto"
	"sync"

	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	sharedMu        sync.RWMutex
	discordSession  *discordgo.Session
	telegramSession *tgbotapi.BotAPI
)

// SetDiscordSession stores the shared Discord session for reuse.
func SetDiscordSession(session *discordgo.Session) {
	sharedMu.Lock()
	defer sharedMu.Unlock()
	discordSession = session
}

// GetDiscordSession returns the shared Discord session.
func GetDiscordSession() *discordgo.Session {
	sharedMu.RLock()
	defer sharedMu.RUnlock()
	return discordSession
}

// SetTelegramBot stores the shared Telegram bot for reuse.
func SetTelegramBot(bot *tgbotapi.BotAPI) {
	sharedMu.Lock()
	defer sharedMu.Unlock()
	telegramSession = bot
}

// GetTelegramBot returns the shared Telegram bot.
func GetTelegramBot() *tgbotapi.BotAPI {
	sharedMu.RLock()
	defer sharedMu.RUnlock()
	return telegramSession
}

func UploadToOSS(svc oss.OSSService, name, mimeType string, data []byte) (string, error) {
	header := &multipart.FileHeader{
		Filename: name,
		Header:   make(textproto.MIMEHeader),
		Size:     int64(len(data)),
	}
	if mimeType != "" {
		header.Header.Set("Content-Type", mimeType)
	}
	url, _, err := svc.UploadFile(&memFile{Reader: bytes.NewReader(data)}, header)
	return url, err
}

type memFile struct {
	*bytes.Reader
}

func (m *memFile) Close() error                                  { return nil }
func (m *memFile) ReadAt(p []byte, off int64) (n int, err error) { return m.Reader.ReadAt(p, off) }
func (m *memFile) Seek(offset int64, whence int) (int64, error)  { return m.Reader.Seek(offset, whence) }
