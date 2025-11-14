package model

// Config 全局配置结构 - 支持点号访问
type Config struct {
	// 环境变量配置
	Port          string
	ListenAddress string
	WebPanelUser  string
	WebPanelPwd   string
	ConfigPath    string

	// 系统配置 - 使用小写字段名，通过 Safe 和 Data 访问
	Safe SafeConfig
	Data DataConfig

	// 友链配置
	FriendLinks []FriendWebsite
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

// FriendWebsite 单个友链站点
type FriendWebsite struct {
	Name   string `json:"name"`
	Link   string `json:"link"`
	Avatar string `json:"avatar"`
	Info   string `json:"info"`
}
