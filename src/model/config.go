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
	Safe    SafeConfig
	Data    DataConfig
	Crawler CrawlerConfig

	// 友链配置
	FriendLinks []FriendWebsite
}

// CrawlerConfig 爬虫配置
type CrawlerConfig struct {
	Concurrency int // 并发数量，默认 5
}

// FriendLinksConf 对应 friend_list.json 的结构
type FriendLinksConf struct {
	FriendLinksData struct {
		Website []FriendWebsite `json:"website"`
	} `json:"friend_links_conf"`
}

// SafeConfig 安全配置
type SafeConfig struct {
	CorsAllowHostlist []string
	ExcludePaths      []string
}

// DataConfig 数据配置
type DataConfig struct {
	Database DatabaseConfig
	Image    ImageConfig
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Path string
}

// ImageConfig 图片配置
type ImageConfig struct {
	Path   string
	ConvTo string
}
