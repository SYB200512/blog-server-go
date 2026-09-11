package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Config 全局配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	AI       AIConfig       `yaml:"ai"`
	Log      LogConfig      `yaml:"log"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	DBName       string `yaml:"dbname"`
	Charset      string `yaml:"charset"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type AIConfig struct {
	DashScopeAPIKey string `yaml:"dashscope_api_key"`
	DashScopeBase   string `yaml:"dashscope_base"`
	CoversDir       string `yaml:"covers_dir"`
}

type LogConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

// 全局配置实例
var Cfg *Config

// DSN 返回 MySQL 连接串
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.Charset)
}

// InitConfig 加载配置文件
func InitConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	// 使用默认配置作为兜底
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.Database.Charset == "" {
		cfg.Database.Charset = "utf8mb4"
	}
	if cfg.JWT.ExpireHours == 0 {
		cfg.JWT.ExpireHours = 24
	}
	if cfg.AI.DashScopeBase == "" {
		cfg.AI.DashScopeBase = "https://dashscope.aliyuncs.com"
	}
	if cfg.AI.CoversDir == "" {
		cfg.AI.CoversDir = "./covers"
	}
	// 支持用环境变量覆盖敏感配置，避免把密码/密钥写进仓库。优先级：环境变量 > config.yaml
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Database.Port = p
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.User = user
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		cfg.Database.DBName = name
	}
	if pwd := os.Getenv("DB_PASSWORD"); pwd != "" {
		cfg.Database.Password = pwd
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.JWT.Secret = secret
	}
	// AI 相关：DASHSCOPE_API_KEY 等优先用环境变量，避免把 key 写进仓库
	if key := os.Getenv("DASHSCOPE_API_KEY"); key != "" {
		cfg.AI.DashScopeAPIKey = key
	}
	if base := os.Getenv("DASHSCOPE_BASE"); base != "" {
		cfg.AI.DashScopeBase = base
	}
	if dir := os.Getenv("COVERS_DIR"); dir != "" {
		cfg.AI.CoversDir = dir
	}
	Cfg = cfg
	return cfg, nil
}

// InitDB 初始化数据库连接
func InitDB(cfg *Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	return db, nil
}
