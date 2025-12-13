package model

import "time"

// Config 全局配置结构 - 支持点号访问
type Config struct {
	// 环境变量配置
	Port              string
	ListenAddress     string
	WebPanelUser      string
	WebPanelPwd       string
	ConfigPath        string
	CronScanOnStartup bool

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

// FriendWebsite 单个友链站点
type FriendWebsite struct {
	ID        int       `json:"id,omitempty" gorm:"column:id;primaryKey"`
	Name      string    `json:"name" gorm:"column:website_name"`
	Link      string    `json:"link" gorm:"column:website_url"`
	Avatar    string    `json:"avatar" gorm:"column:website_icon_url"`
	Info      string    `json:"description" gorm:"column:description"`
	Email     string    `json:"email,omitempty" gorm:"column:email"`
	Times     int       `json:"times,omitempty" gorm:"column:times"`
	Status    string    `json:"status,omitempty" gorm:"column:status"`
	EnableRss bool      `json:"enable_rss,omitempty" gorm:"column:enable_rss"`
	UpdatedAt time.Time `json:"updated_at,omitempty" gorm:"column:updated_at"`
}

// TableName sets the insert table name for this struct type.
func (FriendWebsite) TableName() string {
	return "friend_link"
}
