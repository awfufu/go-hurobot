package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"slices"

	"gopkg.in/yaml.v3"
)

// Config 配置结构体
type yamlConfig struct {
	// NapCat 配置
	HttpRemote string `yaml:"http_remote"` // 正向 HTTP 地址
	HttpListen string `yaml:"http_listen"` // 反向 HTTP 监听端口

	// API Keys 配置
	ApiKeys struct {
		DrawUrlBase        string `yaml:"draw_url_base"`
		DrawApiKey         string `yaml:"draw_api_key"`
		ExchangeRateAPIKey string `yaml:"exchange_rate_api_key"`
		OkxMirrorAPIKey    string `yaml:"okx_mirror_api_key"`
	} `yaml:"api_keys"`

	// SQLite 配置
	SQLite struct {
		Path string `yaml:"path"`
	} `yaml:"sqlite"`

	// 权限配置
	Permissions struct {
		MasterID         uint64   `yaml:"master_id"`
		BotID            uint64   `yaml:"bot_id"`
		AdminIDs         []uint64 `yaml:"admin_ids,omitempty"`
		BotOwnerGroupIDs []uint64 `yaml:"bot_owner_group_ids,omitempty"`
	} `yaml:"permissions"`

	// 其他配置
	ProxyURL  string                    `yaml:"proxy_url,omitempty"`
	Suppliers map[string]SupplierConfig `yaml:"suppliers,omitempty"`
}

type SupplierConfig struct {
	BaseURL      string `yaml:"base_url"`
	APIKey       string `yaml:"api_key"`
	DefaultModel string `yaml:"default_model"`
	Proxy        string `yaml:"proxy,omitempty"`
}

type Permission int

const (
	Guest  Permission = 0 // 所有人都可以使用
	Admin  Permission = 3 // 管理员及以上
	Master Permission = 4 // 仅 Master
)

var Cfg yamlConfig

func LoadConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := yaml.Unmarshal(data, &Cfg); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	// NapCat 默认值
	if Cfg.HttpRemote == "" {
		Cfg.HttpRemote = "http://127.0.0.1:3000"
	}
	if Cfg.HttpListen == "" {
		Cfg.HttpListen = "0.0.0.0:3001"
	}

	// 权限默认值
	if Cfg.Permissions.MasterID == 0 {
		Cfg.Permissions.MasterID = 1006554341
	}
	if Cfg.Permissions.BotID == 0 {
		Cfg.Permissions.BotID = 3552586437
	}

	// SQLite 默认值
	if Cfg.SQLite.Path == "" {
		Cfg.SQLite.Path = "db/bot.db"
	}

	return nil
}

func LoadConfigFile() {
	configPathPtr := flag.String("c", "config.yaml", "配置文件路径")
	flag.Parse()

	configPath = *configPathPtr
	if err := LoadConfig(configPath); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
}

// GetConfigPath returns the config path
func GetConfigPath() string {
	return configPath
}

func GetUserPermission(userID uint64) Permission {
	if userID == Cfg.Permissions.MasterID {
		return Master
	}
	if Cfg.IsAdmin(userID) {
		return Admin
	}
	return Guest
}

// IsAdmin 检查用户是否是管理员
func (cfg *yamlConfig) IsAdmin(userID uint64) bool {
	if userID == cfg.Permissions.MasterID {
		return true
	}
	return slices.Contains(cfg.Permissions.AdminIDs, userID)
}

func AddAdmin(userID uint64) error {
	if userID == Cfg.Permissions.MasterID {
		return fmt.Errorf("user is already master")
	}
	if slices.Contains(Cfg.Permissions.AdminIDs, userID) {
		return fmt.Errorf("user is already admin")
	}
	Cfg.Permissions.AdminIDs = append(Cfg.Permissions.AdminIDs, userID)
	return SaveConfig()
}

func RemoveAdmin(userID uint64) error {
	if userID == Cfg.Permissions.MasterID {
		return fmt.Errorf("cannot remove master")
	}
	idx := -1
	for i, id := range Cfg.Permissions.AdminIDs {
		if id == userID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("user is not admin")
	}
	Cfg.Permissions.AdminIDs = append(Cfg.Permissions.AdminIDs[:idx], Cfg.Permissions.AdminIDs[idx+1:]...)
	return SaveConfig()
}

func GetAdmins() []uint64 {
	return Cfg.Permissions.AdminIDs
}

var configPath string

func SetConfigPath(path string) {
	configPath = path
}

func SaveConfig() error {
	if configPath == "" {
		configPath = "config.yaml"
	}
	data, err := yaml.Marshal(&Cfg)
	if err != nil {
		return fmt.Errorf("marshal config failed: %w", err)
	}
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("write config file failed: %w", err)
	}
	return nil
}

func ReloadConfig() error {
	if configPath == "" {
		configPath = "config.yaml"
	}
	return LoadConfig(configPath)
}
