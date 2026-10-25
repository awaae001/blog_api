package config

import (
	"blog_api/src/model"
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

//go:embed defaults/*.json
var defaultConfigFiles embed.FS

var (
	currentConfig   atomic.Pointer[model.Config]
	configMu        sync.Mutex
	configValidator = validator.New(validator.WithRequiredStructEnabled())
)

type UpdateResult struct {
	Message             string   `json:"message"`
	Reloaded            bool     `json:"reloaded"`
	RestartRequiredKeys []string `json:"restart_required_keys"`
}

// Load 初始化配置；失败后可以重试。
func Load() (*model.Config, error) {
	configMu.Lock()
	defer configMu.Unlock()
	if cfg := currentConfig.Load(); cfg != nil {
		return cfg, nil
	}
	return reload()
}

// GetConfig 在 Load 后返回只读快照，已取得的快照不会随 Reload 改变。
func GetConfig() *model.Config {
	return currentConfig.Load()
}

// Reload 重新加载配置；失败时保留上一份运行时快照。
func Reload() (*model.Config, error) {
	configMu.Lock()
	defer configMu.Unlock()
	return reload()
}

func reload() (*model.Config, error) {
	cfg, err := loadConfig(nil)
	if err != nil {
		return nil, err
	}
	publish(cfg)
	return cfg, nil
}

// loadConfig 只构建、校验候选配置，不发布。data 为 nil 时读取磁盘文件。
func loadConfig(data []byte) (*model.Config, error) {
	v := viper.NewWithOptions(viper.ExperimentalBindStruct())
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	for key, value := range map[string]any{
		"PORT": "10024", "LISTEN_ADDRESS": "0.0.0.0",
		"CONFIG_PATH": "data/config", "CRON_SCAN_ON_STARTUP": true,
	} {
		v.SetDefault(key, value)
	}
	if err := v.ReadInConfig(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("读取 .env 失败: %w", err)
	}
	active := currentConfig.Load()
	if active != nil {
		v.Set("CONFIG_PATH", active.ConfigPath)
	}
	configPath := v.GetString("CONFIG_PATH")
	if active == nil {
		if err := ensureDefaultConfigFiles(configPath); err != nil {
			return nil, err
		}
	}
	if data == nil {
		var err error
		data, err = os.ReadFile(filepath.Join(configPath, "system_config.json"))
		if err != nil {
			return nil, err
		}
	}
	v.SetConfigType("json")
	if err := v.MergeConfig(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("解析 system_config.json 失败: %w", err)
	}
	for env, key := range map[string]string{
		"TELEGRAM_BOT_TOKEN":    "moments_integrated_conf.integrated.telegram.bot_token",
		"DISCORD_BOT_TOKEN":     "moments_integrated_conf.integrated.discord.bot_token",
		"OSS_ACCESS_KEY_ID":     "oss_conf.accessKeyId",
		"OSS_ACCESS_KEY_SECRET": "oss_conf.accessKeySecret",
		"TURNSTILE_SECRET":      "verify_conf.turnstile.secret",
		"FINGERPRINT_SECRET":    "verify_conf.fingerprint.secret",
		"EMAIL_PASSWORD":        "email_conf.password",
	} {
		if value := v.GetString(env); value != "" {
			v.Set("system_conf."+key, value)
		}
	}
	cfg := new(model.Config)
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	if cfg.Data.Database.MigrationsPath == "" {
		cfg.Data.Database.MigrationsPath = "migrations"
	}
	cfg.Data.Resource.Path = strings.TrimSpace(cfg.Data.Resource.Path)
	if cfg.Data.Resource.Path == "" {
		cfg.Data.Resource.Path = "data"
	}
	cfg.Data.Resource.Path = filepath.Clean(cfg.Data.Resource.Path)
	cfg.Data.Image.ConvTo = strings.ToLower(strings.TrimSpace(cfg.Data.Image.ConvTo))
	if cfg.Crawler.Concurrency <= 0 {
		cfg.Crawler.Concurrency = 5
	}
	if cfg.Crawler.RssTimeoutSeconds <= 0 {
		cfg.Crawler.RssTimeoutSeconds = 15
	}
	cfg.Safe.EnabledPublicAPIs = normalizeEnabledPublicAPIs(cfg.Safe.EnabledPublicAPIs)
	targets := &cfg.MomentsIntegrated.Integrated
	targets.Telegram.BotToken = strings.TrimSpace(targets.Telegram.BotToken)
	targets.Discord.BotToken = strings.TrimSpace(targets.Discord.BotToken)
	cfg.Verify.Turnstile.Secret = strings.TrimSpace(cfg.Verify.Turnstile.Secret)
	cfg.Verify.Fingerprint.Secret = strings.TrimSpace(cfg.Verify.Fingerprint.Secret)
	if cfg.Email.Sender == "" {
		cfg.Email.Sender = cfg.Email.UserName
	}
	if err := configValidator.Struct(cfg); err != nil {
		return nil, fmt.Errorf("配置校验失败: %w", err)
	}
	if active != nil {
		return cfg, nil
	}
	friendData, err := os.ReadFile(filepath.Join(configPath, "friend_list.json"))
	if err != nil {
		return nil, err
	}
	var friends model.FriendLinksConf
	if err := json.Unmarshal(friendData, &friends); err != nil {
		return nil, fmt.Errorf("解析 friend_list.json 失败: %w", err)
	}
	cfg.FriendLinks = friends.FriendLinksData.Website
	return cfg, nil
}

// publish 保留启动型配置，避免快照与现有数据库、客户端使用不同的参数。
func publish(cfg *model.Config) {
	if active := currentConfig.Load(); active != nil {
		cfg.EnvironmentConfig = active.EnvironmentConfig
		cfg.Data.Database = active.Data.Database
		cfg.Data.Image = active.Data.Image
		cfg.MomentsIntegrated = active.MomentsIntegrated
		cfg.OSS = active.OSS
		cfg.FriendLinks = active.FriendLinks
	}
	cfg.Safe.ExcludePaths = append(cfg.Safe.ExcludePaths, cfg.Data.Database.Path)
	currentConfig.Store(cfg)
}

func ensureDefaultConfigFiles(configPath string) error {
	if err := os.MkdirAll(configPath, 0o755); err != nil {
		return err
	}
	for _, name := range []string{"system_config.json", "friend_list.json"} {
		content, err := defaultConfigFiles.ReadFile("defaults/" + name)
		if err != nil {
			return err
		}
		path := filepath.Join(configPath, name)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return err
		}
		_, writeErr := file.Write(content)
		if err := errors.Join(writeErr, file.Close()); err != nil {
			_ = os.Remove(path)
			return fmt.Errorf("写入默认配置 %s 失败: %w", path, err)
		}
	}
	return nil
}

func normalizeEnabledPublicAPIs(values []string) []string {
	allowed := []string{model.PublicAPIMoments, model.PublicAPIImage, model.PublicAPIFriend, model.PublicAPIRSS, model.PublicAPIEmail}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if !slices.Contains(allowed, value) {
			log.Printf("[config][WARN]忽略未知的公开 API 功能键: %q", value)
		} else if !slices.Contains(result, value) {
			result = append(result, value)
		}
	}
	return result
}

// UpdateAndSaveConfig 按 RFC 7396 合并局部配置，校验后原子落盘并发布。
func UpdateAndSaveConfig(patch model.UpdateConfigReq) (*UpdateResult, error) {
	configMu.Lock()
	defer configMu.Unlock()

	active := GetConfig()
	filePath := filepath.Join(active.ConfigPath, "system_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	patchData, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}
	data, err = jsonpatch.MergePatch(data, patchData)
	if err != nil {
		return nil, err
	}
	data, err = json.MarshalIndent(json.RawMessage(data), "", "  ")
	if err != nil {
		return nil, err
	}
	cfg, err := loadConfig(data)
	if err != nil {
		return nil, err
	}
	result := &UpdateResult{
		Message:             "配置更新成功，已刷新运行时配置",
		Reloaded:            true,
		RestartRequiredKeys: restartRequiredKeys(active, cfg),
	}
	if err := writeFileAtomic(filePath, data); err != nil {
		return nil, fmt.Errorf("写入 system_config.json 失败: %w", err)
	}
	publish(cfg)
	return result, nil
}

func writeFileAtomic(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".system_config-*.tmp")
	if err != nil {
		return err
	}
	defer file.Close()
	defer os.Remove(file.Name())
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

func restartRequiredKeys(active, next *model.Config) []string {
	keys := make([]string, 0)
	for _, field := range []struct {
		key     string
		changed bool
	}{
		{"data_conf.database", active.Data.Database != next.Data.Database},
		{"data_conf.image.path", active.Data.Image.Path != next.Data.Image.Path},
		{"data_conf.image.conv_to", active.Data.Image.ConvTo != next.Data.Image.ConvTo},
		{"moments_integrated_conf", !reflect.DeepEqual(active.MomentsIntegrated, next.MomentsIntegrated)},
		{"oss_conf", active.OSS != next.OSS},
	} {
		if field.changed {
			keys = append(keys, "system_conf."+field.key)
		}
	}
	return keys
}
