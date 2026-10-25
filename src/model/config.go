package model

// Config 是只读的运行时配置快照。
type Config struct {
	EnvironmentConfig `mapstructure:",squash"`
	SystemConfig      `mapstructure:"system_conf"`
	FriendLinks       []FriendWebsite `mapstructure:"-"`
}

// EnvironmentConfig 在启动时生效，重载不会切换应用的启动参数。
type EnvironmentConfig struct {
	Port                   string `mapstructure:"PORT"`
	ListenAddress          string `mapstructure:"LISTEN_ADDRESS"`
	WebPanelUser           string `mapstructure:"WEB_PANEL_USER"`
	WebPanelPwd            string `mapstructure:"WEB_PANEL_PWD"`
	StateAPIMasterPassword string `mapstructure:"STATE_API_MASTER_PASSWORD"`
	ConfigPath             string `mapstructure:"CONFIG_PATH"`
	CronScanOnStartup      bool   `mapstructure:"CRON_SCAN_ON_STARTUP"`
	IsDev                  bool   `mapstructure:"IS_DEV"`
}

// SystemConfig 对应 system_config.json 的 system_conf。
type SystemConfig struct {
	Safe              SafeConfig              `mapstructure:"safe_conf"`
	Data              DataConfig              `mapstructure:"data_conf"`
	Crawler           CrawlerConfig           `mapstructure:"crawler_conf"`
	MomentsIntegrated MomentsIntegratedConfig `mapstructure:"moments_integrated_conf"`
	OSS               OSSConfig               `mapstructure:"oss_conf"`
	Verify            VerifyConfig            `mapstructure:"verify_conf"`
	Email             EmailConf               `mapstructure:"email_conf"`
}

const (
	PublicAPIMoments = "moments"
	PublicAPIImage   = "image"
	PublicAPIFriend  = "friend"
	PublicAPIRSS     = "rss"
	PublicAPIEmail   = "email"
)

// CrawlerConfig 爬虫配置
type CrawlerConfig struct {
	Concurrency       int `mapstructure:"concurrency"`         // 并发数量，默认 5
	RssTimeoutSeconds int `mapstructure:"rss_timeout_seconds"` // RSS 解析超时（秒）
}

// FriendLinksConf 对应 friend_list.json 的结构
type FriendLinksConf struct {
	FriendLinksData struct {
		Website []FriendWebsite `json:"website"`
	} `json:"friend_links_conf"`
}

// SafeConfig 安全配置
type SafeConfig struct {
	CorsAllowHostlist []string `mapstructure:"cors_allow_hostlist"`
	ExcludePaths      []string `mapstructure:"exclude_paths"`
	AllowExtension    []string `mapstructure:"allow_extension"`
	EnabledPublicAPIs []string `mapstructure:"enabled_public_apis"`
}

// IsPublicAPIEnabled reports whether a normalized public API key is enabled.
func (c SafeConfig) IsPublicAPIEnabled(key string) bool {
	for _, enabled := range c.EnabledPublicAPIs {
		if enabled == key {
			return true
		}
	}
	return false
}

// DataConfig 数据配置
type DataConfig struct {
	Database DatabaseConfig `mapstructure:"database"`
	Image    ImageConfig    `mapstructure:"image"`
	Resource ResourceConfig `mapstructure:"resource"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Path           string `mapstructure:"path" validate:"required"`
	MigrationsPath string `mapstructure:"migrations_path"`
}

// ImageConfig 图片配置
type ImageConfig struct {
	Path   string `mapstructure:"path"`
	ConvTo string `mapstructure:"conv_to" validate:"omitempty,oneof=webp jpg jpeg png gif"`
}

// ResourceConfig 资源配置
type ResourceConfig struct {
	Path string `mapstructure:"path"`
}

// MomentsIntegratedConfig 动态集成配置
type MomentsIntegratedConfig struct {
	Enable                 bool              `mapstructure:"enable"`
	ApiSingleReturnEntries int               `mapstructure:"api_single_return_entries"`
	Integrated             IntegratedTargets `mapstructure:"integrated"`
}

// OSSConfig OSS 配置
type OSSConfig struct {
	Provider        string `mapstructure:"provider" validate:"required_if=Enable true,omitempty,oneof=aliyun s3"`
	Enable          bool   `mapstructure:"enable"`
	AccessKeyID     string `mapstructure:"accessKeyId" validate:"required_if=Enable true"`
	AccessKeySecret string `mapstructure:"accessKeySecret" validate:"required_if=Enable true"`
	Bucket          string `mapstructure:"bucket" validate:"required_if=Enable true"`
	Endpoint        string `mapstructure:"endpoint" validate:"required_if=Enable true Provider aliyun"`
	Region          string `mapstructure:"region" validate:"required_if=Enable true Provider s3"`
	CustomDomain    string `mapstructure:"customDomain"`
	Secure          bool   `mapstructure:"secure"`
	Timeout         int    `mapstructure:"timeout"`
	Prefix          string `mapstructure:"prefix"`
}

// VerifyConfig 验证配置
type VerifyConfig struct {
	Turnstile   TurnstileConfig   `mapstructure:"turnstile"`
	Fingerprint FingerprintConfig `mapstructure:"fingerprint"`
}

// TurnstileConfig Turnstile 配置
type TurnstileConfig struct {
	Enable  bool   `mapstructure:"enable"`
	Secret  string `mapstructure:"secret" validate:"required_if=Enable true"`
	SiteKey string `mapstructure:"site_key"`
}

// FingerprintConfig 指纹配置
type FingerprintConfig struct {
	Secret string `mapstructure:"secret" validate:"required"`
}

// IntegratedTargets 集成目标
type IntegratedTargets struct {
	Telegram TelegramConfig `mapstructure:"telegram"`
	Discord  DiscordConfig  `mapstructure:"discord"`
}

// TelegramConfig Telegram 配置
type TelegramConfig struct {
	Enable       bool     `mapstructure:"enable"`
	SyncDelete   bool     `mapstructure:"sync_delete"`
	BotToken     string   `mapstructure:"bot_token"`
	ChannelID    string   `mapstructure:"channel_id"`
	FilterUserid []string `mapstructure:"filter_userid"`
	StartTime    int64    `mapstructure:"start_time"` // Unix timestamp, 只接受此时间之后的消息（0 表示使用 bot 启动时间）
}

// EmailConf 邮箱配置
type EmailConf struct {
	Enable   bool   `mapstructure:"enable"`
	Host     string `mapstructure:"host" validate:"required_if=Enable true"`
	UserName string `mapstructure:"user_name" validate:"required_if=Enable true"`
	Password string `mapstructure:"password"`
	Port     int    `mapstructure:"port" validate:"required_if=Enable true,omitempty,min=1,max=65535"`
	Sender   string `mapstructure:"sender"`
}

// DiscordConfig Discord 配置
type DiscordConfig struct {
	Enable       bool     `mapstructure:"enable"`
	SyncDelete   bool     `mapstructure:"sync_delete"`
	BotToken     string   `mapstructure:"bot_token"`
	GuildID      string   `mapstructure:"guild_id"`
	ChannelID    string   `mapstructure:"channel_id"`
	FilterUserid []string `mapstructure:"filter_userid"`
}
