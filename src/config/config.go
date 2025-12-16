package config

import (
	"blog_api/src/model"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

var (
	globalConfig *model.Config
	once         sync.Once
	v            *viper.Viper // 全局 viper 实例
)

// Load 加载所有配置 (单例模式)
// 配置加载顺序:
// 1. .env 文件 (环境变量)
// 2. data/config/ 目录下的 JSON 文件 (通过 MergeInConfig 合并)
// 环境变量会覆盖配置文件中的同名设置
func Load() (*model.Config, error) {
	var err error
	once.Do(func() {
		globalConfig, err = loadConfig()
	})
	return globalConfig, err
}

// GetConfig 获取全局配置实例
func GetConfig() *model.Config {
	if globalConfig == nil {
		log.Fatal("配置未初始化,请先调用 Load()")
	}
	return globalConfig
}

// loadConfig 执行实际的配置加载
func loadConfig() (*model.Config, error) {
	// 初始化全局 viper 实例
	v = viper.New()

	// 设置默认值
	v.SetDefault("CRON_SCAN_ON_STARTUP", true)
	v.SetDefault("ENABLE_STATUS_LOG", false)

	// 1. 从 .env 文件加载环境变量
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()                                   // 自动读取匹配的环境变量
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // 将配置键中的'.'替换为'_'以匹配环境变量

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("未找到 .env 文件，将跳过加载")
		} else {
			return nil, fmt.Errorf("解析 .env 文件时发生错误: %w", err)
		}
	}

	// 2. 获取配置路径 (从环境变量或使用默认值)
	configPath := v.GetString("CONFIG_PATH")
	if configPath == "" {
		configPath = "data/config"
	}

	// 3. 合并 system_config.json
	if err := mergeJSONConfig("system_config", configPath); err != nil {
		return nil, err
	}

	// 4. 合并 friend_list.json
	if err := mergeJSONConfig("friend_list", configPath); err != nil {
		return nil, err
	}

	// 5. 解析配置到结构体
	cfg := &model.Config{}
	if err := unmarshalConfig(cfg); err != nil {
		return nil, fmt.Errorf("解析配置到结构体失败: %w", err)
	}

	return cfg, nil
}

// mergeJSONConfig 合并指定的 JSON 配置文件
func mergeJSONConfig(configName, configPath string) error {
	v.SetConfigName(configName)
	v.SetConfigType("json")
	v.AddConfigPath(configPath)

	if err := v.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("未找到配置文件 (%s/%s.json)，将跳过合并", configPath, configName)
			return nil
		}
		return fmt.Errorf("合并配置文件 %s 时发生错误: %w", configName, err)
	}

	log.Printf("已合并配置文件: %s/%s.json", configPath, configName)
	return nil
}

// unmarshalConfig 将 viper 配置解析到 Config 结构体
func unmarshalConfig(cfg *model.Config) error {
	// 解析环境变量
	cfg.Port = v.GetString("PORT")
	cfg.ListenAddress = v.GetString("LISTEN_ADDRESS")
	cfg.WebPanelUser = v.GetString("WEB_PANEL_USER")
	cfg.WebPanelPwd = v.GetString("WEB_PANEL_PWD")
	cfg.ConfigPath = v.GetString("CONFIG_PATH")
	cfg.CronScanOnStartup = v.GetBool("CRON_SCAN_ON_STARTUP")
	cfg.EnableStatusLog = v.GetBool("ENABLE_STATUS_LOG")

	// 设置默认值
	if cfg.Port == "" {
		cfg.Port = "10024"
	}
	if cfg.ListenAddress == "" {
		cfg.ListenAddress = "0.0.0.0"
	}
	if cfg.ConfigPath == "" {
		cfg.ConfigPath = "data/config"
	}

	// 解析系统配置
	cfg.Safe.CorsAllowHostlist = v.GetStringSlice("system_conf.safe_conf.cors_allow_hostlist")
	cfg.Safe.ExcludePaths = v.GetStringSlice("system_conf.safe_conf.exclude_paths")
	cfg.Safe.AllowExtension = v.GetStringSlice("system_conf.safe_conf.allow_extension")
	cfg.Data.Database.Path = v.GetString("system_conf.data_conf.database.path")
	cfg.Data.Image.Path = v.GetString("system_conf.data_conf.image.path")
	cfg.Data.Image.ConvTo = v.GetString("system_conf.data_conf.image.conv_to")
	cfg.Data.Resource.Path = v.GetString("system_conf.data_conf.resource.path")

	// 动态地将核心数据路径添加到排除列表，以防止被意外删除
	if cfg.Data.Database.Path != "" {
		cfg.Safe.ExcludePaths = append(cfg.Safe.ExcludePaths, cfg.Data.Database.Path)
	}

	// 解析爬虫配置
	cfg.Crawler.Concurrency = v.GetInt("system_conf.crawler_conf.concurrency")
	if cfg.Crawler.Concurrency <= 0 {
		cfg.Crawler.Concurrency = 5 // 默认并发数为 5
	}

	// 手动解析友链配置
	friendListPath := filepath.Join(cfg.ConfigPath, "friend_list.json")
	friendListData, err := ioutil.ReadFile(friendListPath)
	if err != nil {
		log.Printf("无法读取 friend_list.json 文件: %v, 将跳过加载友链", err)
	} else {
		var friendLinksConf model.FriendLinksConf
		if err := json.Unmarshal(friendListData, &friendLinksConf); err != nil {
			return fmt.Errorf("解析 friend_list.json 文件失败: %w", err)
		}
		cfg.FriendLinks = friendLinksConf.FriendLinksData.Website
	}

	return nil
}

// Reload 重新加载配置 (用于热更新)
func Reload() error {
	once = sync.Once{} // 重置 once，允许重新加载
	newConfig, err := loadConfig()
	if err != nil {
		return err
	}
	globalConfig = newConfig
	log.Println("配置已重新加载")
	return nil
}
