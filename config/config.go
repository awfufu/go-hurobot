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
	NapcatHttpServer  string `yaml:"napcat_http_server"`  // 正向 HTTP 地址
	ReverseHttpListen string `yaml:"reverse_http_listen"` // 反向 HTTP 监听端口

	// API Keys 配置
	ApiKeys struct {
		DrawUrlBase        string `yaml:"draw_url_base"`
		DrawApiKey         string `yaml:"draw_api_key"`
		ExchangeRateAPIKey string `yaml:"exchange_rate_api_key"`
		OkxMirrorAPIKey    string `yaml:"okx_mirror_api_key"`
		Longport           struct {
			AppKey          string `yaml:"app_key"`
			AppSecret       string `yaml:"app_secret"`
			AccessToken     string `yaml:"access_token"`
			Region          string `yaml:"region"`
			EnableOvernight bool   `yaml:"enable_overnight"`
		} `yaml:"longport"`
	} `yaml:"api_keys"`

	// PostgreSQL 配置
	PostgreSQL struct {
		Host     string `yaml:"host"`
		Port     uint16 `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DbName   string `yaml:"dbname"`
	} `yaml:"postgresql"`

	// Python 配置
	Python struct {
		Interpreter string `yaml:"interpreter"`
	} `yaml:"python"`

	// 权限配置
	Permissions struct {
		MasterID         uint64      `yaml:"master_id"`
		BotID            uint64      `yaml:"bot_id"`
		AdminIDs         []uint64    `yaml:"admin_ids,omitempty"`
		BotOwnerGroupIDs []uint64    `yaml:"bot_owner_group_ids,omitempty"`
		Cmds             []CmdConfig `yaml:"cmds,omitempty"`
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

// CmdConfig 命令配置结构
type CmdConfig struct {
	Name        string   `yaml:"name"`                   // 命令名称
	Permission  string   `yaml:"permission,omitempty"`   // 权限级别: guest, admin, master，默认为 master
	AllowUsers  []uint64 `yaml:"allow_users,omitempty"`  // 允许执行的用户列表
	RejectUsers []uint64 `yaml:"reject_users,omitempty"` // 拒绝执行的用户列表
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
	if Cfg.NapcatHttpServer == "" {
		Cfg.NapcatHttpServer = "http://127.0.0.1:3000"
	}
	if Cfg.ReverseHttpListen == "" {
		Cfg.ReverseHttpListen = "0.0.0.0:3001"
	}

	// 权限默认值
	if Cfg.Permissions.MasterID == 0 {
		Cfg.Permissions.MasterID = 1006554341
	}
	if Cfg.Permissions.BotID == 0 {
		Cfg.Permissions.BotID = 3552586437
	}

	// PostgreSQL 默认值
	if Cfg.PostgreSQL.Host == "" {
		Cfg.PostgreSQL.Host = "127.0.0.1"
	}
	if Cfg.PostgreSQL.Port == 0 {
		Cfg.PostgreSQL.Port = 5432
	}

	// Longport 默认值
	if Cfg.ApiKeys.Longport.Region == "" {
		Cfg.ApiKeys.Longport.Region = "cn"
	}

	// Python 默认值
	if Cfg.Python.Interpreter == "" {
		Cfg.Python.Interpreter = "python3"
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

	log.Printf("配置加载成功: %s", configPath)
}

// GetCmdConfig 获取指定命令的配置
func (cfg *yamlConfig) GetCmdConfig(cmdName string) *CmdConfig {
	for i := range cfg.Permissions.Cmds {
		if cfg.Permissions.Cmds[i].Name == cmdName {
			return &cfg.Permissions.Cmds[i]
		}
	}
	return nil
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

// IsInAllowList 检查用户是否在命令的允许列表中
func (c *CmdConfig) IsInAllowList(userID uint64) bool {
	if len(c.AllowUsers) == 0 {
		return false
	}
	for _, id := range c.AllowUsers {
		if id == userID {
			return true
		}
	}
	return false
}

// IsInRejectList 检查用户是否在命令的拒绝列表中
func (c *CmdConfig) IsInRejectList(userID uint64) bool {
	if len(c.RejectUsers) == 0 {
		return false
	}
	for _, id := range c.RejectUsers {
		if id == userID {
			return true
		}
	}
	return false
}

// GetPermissionLevel 获取权限级别（返回数值便于比较）
func (c *CmdConfig) GetPermissionLevel() Permission {
	switch c.Permission {
	case "guest":
		return 0
	case "admin":
		return 3
	case "master":
		return 4
	default:
		return 4 // 默认为 master
	}
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

func AddAllowUser(cmdName string, userID uint64) error {
	cmdCfg := Cfg.GetCmdConfig(cmdName)
	if cmdCfg == nil {
		cmdCfg = &CmdConfig{Name: cmdName}
		Cfg.Permissions.Cmds = append(Cfg.Permissions.Cmds, *cmdCfg)
		cmdCfg = &Cfg.Permissions.Cmds[len(Cfg.Permissions.Cmds)-1]
	}
	if slices.Contains(cmdCfg.AllowUsers, userID) {
		return fmt.Errorf("user already in allow list")
	}
	cmdCfg.AllowUsers = append(cmdCfg.AllowUsers, userID)
	return SaveConfig()
}

func RemoveAllowUser(cmdName string, userID uint64) error {
	cmdCfg := Cfg.GetCmdConfig(cmdName)
	if cmdCfg == nil {
		return fmt.Errorf("command config not found")
	}
	idx := -1
	for i, id := range cmdCfg.AllowUsers {
		if id == userID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("user not in allow list")
	}
	cmdCfg.AllowUsers = append(cmdCfg.AllowUsers[:idx], cmdCfg.AllowUsers[idx+1:]...)
	return SaveConfig()
}

func GetAllowUsers(cmdName string) ([]uint64, error) {
	cmdCfg := Cfg.GetCmdConfig(cmdName)
	if cmdCfg == nil {
		return []uint64{}, nil
	}
	return cmdCfg.AllowUsers, nil
}

func AddRejectUser(cmdName string, userID uint64) error {
	cmdCfg := Cfg.GetCmdConfig(cmdName)
	if cmdCfg == nil {
		cmdCfg = &CmdConfig{Name: cmdName}
		Cfg.Permissions.Cmds = append(Cfg.Permissions.Cmds, *cmdCfg)
		cmdCfg = &Cfg.Permissions.Cmds[len(Cfg.Permissions.Cmds)-1]
	}
	if slices.Contains(cmdCfg.RejectUsers, userID) {
		return fmt.Errorf("user already in reject list")
	}
	cmdCfg.RejectUsers = append(cmdCfg.RejectUsers, userID)
	return SaveConfig()
}

func RemoveRejectUser(cmdName string, userID uint64) error {
	cmdCfg := Cfg.GetCmdConfig(cmdName)
	if cmdCfg == nil {
		return fmt.Errorf("command config not found")
	}
	idx := -1
	for i, id := range cmdCfg.RejectUsers {
		if id == userID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("user not in reject list")
	}
	cmdCfg.RejectUsers = append(cmdCfg.RejectUsers[:idx], cmdCfg.RejectUsers[idx+1:]...)
	return SaveConfig()
}

func GetRejectUsers(cmdName string) ([]uint64, error) {
	cmdCfg := Cfg.GetCmdConfig(cmdName)
	if cmdCfg == nil {
		return []uint64{}, nil
	}
	return cmdCfg.RejectUsers, nil
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
