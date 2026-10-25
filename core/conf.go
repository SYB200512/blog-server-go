package core

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"

	"blog-backend/config"
	"blog-backend/global"
)

// InitConf 读取 config/config.yaml 初始化全局配置
func InitConf() {
	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("读取配置文件失败: %v", err))
	}
	cfg := &config.Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		panic(fmt.Sprintf("解析配置文件失败: %v", err))
	}
	applyDefaults(cfg)
	overrideWithEnv(cfg)
	global.Config = cfg
}

// applyDefaults 使用默认配置作为兜底
func applyDefaults(cfg *config.Config) {
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
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Director == "" {
		cfg.Log.Director = "logs"
	}
	if cfg.Log.MaxSize == 0 {
		cfg.Log.MaxSize = 10
	}
	if cfg.Log.MaxBackups == 0 {
		cfg.Log.MaxBackups = 7
	}
	if cfg.Log.MaxAge == 0 {
		cfg.Log.MaxAge = 30
	}
}

// overrideWithEnv 用环境变量覆盖敏感配置，避免把密码/密钥写进仓库。优先级：环境变量 > config.yaml
func overrideWithEnv(cfg *config.Config) {
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
}
