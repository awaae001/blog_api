package model

// Config 全局配置结构 - 支持点号访问
type Config struct {
	// 环境变量配置
	Port              string
	ListenAddress     string
	WebPanelUser      string
	WebPanelPwd       string
	ConfigPath        string
	CronScanOnStartup bool
	EnableStatusLog   bool

	// 系统配置 - 使用小写字段名，通过 Safe 和 Data 访问
	Safe    SafeConfig    `mapstructure:"safe_conf"`
	Data    DataConfig    `mapstructure:"data_conf"`
	Crawler CrawlerConfig `mapstructure:"crawler_conf"`

	// 友链配置
	FriendLinks []FriendWebsite
}

// CrawlerConfig 爬虫配置
type CrawlerConfig struct {
	Concurrency int `mapstructure:"concurrency"` // 并发数量，默认 5
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
}

// DataConfig 数据配置
type DataConfig struct {
	Database DatabaseConfig `mapstructure:"database"`
	Image    ImageConfig    `mapstructure:"image"`
	Resource ResourceConfig `mapstructure:"resource"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// ImageConfig 图片配置
type ImageConfig struct {
	Path   string `mapstructure:"path"`
	ConvTo string `mapstructure:"conv_to"`
}

// ResourceConfig 资源配置
type ResourceConfig struct {
	Path string `mapstructure:"path"`
}
