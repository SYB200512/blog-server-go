package core

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"blog-backend/config"
	"blog-backend/global"
)

// 默认配置文件路径；容器内工作目录不固定，可用 CONFIG_PATH 覆盖
const defaultConfigPath = "config/config.yaml"

// InitConf 读取配置并做启动前校验。
// 顺序固定：读取 → applyDefaults → 环境变量覆盖 → Validate → 写入 global.Config。
// warnings 是可选功能降级提示；error 非 nil 表示硬性缺陷，调用方必须退出。
func InitConf() (warnings []string, err error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = defaultConfigPath
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败(%s): %w", path, err)
	}
	cfg := &config.Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败(%s): %w", path, err)
	}
	applyDefaults(cfg)
	// 必须在环境变量覆盖之后校验，否则只用 env 注入的密钥看不到
	if err := overrideWithEnv(cfg); err != nil {
		return nil, err
	}
	warnings, err = cfg.Validate()
	if err != nil {
		return warnings, err
	}
	global.Config = cfg
	return warnings, nil
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
	if cfg.AI.Cover.BaseURL == "" {
		cfg.AI.Cover.BaseURL = "https://dashscope.aliyuncs.com"
	}
	if cfg.AI.Cover.CoversDir == "" {
		cfg.AI.Cover.CoversDir = "./covers"
	}
	if cfg.Storage.Type == "" {
		cfg.Storage.Type = "local"
	}
	if cfg.Storage.MaxSize == 0 {
		cfg.Storage.MaxSize = 5
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

	// P0：HTTP Server 超时默认值，避免 yaml 漏配时出现 0 值
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 60
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 60
	}
	if cfg.Server.IdleTimeout == 0 {
		cfg.Server.IdleTimeout = 75
	}
	if cfg.Server.MaxHeaderBytes == 0 {
		cfg.Server.MaxHeaderBytes = 1 << 20
	}
	if cfg.Server.PprofAddr == "" {
		cfg.Server.PprofAddr = "127.0.0.1:6060" // 填 "off" 可关闭
	}
	// P0：连接池生命周期默认值
	if cfg.Database.ConnMaxLifetime == 0 {
		cfg.Database.ConnMaxLifetime = 30
	}
	if cfg.Database.ConnMaxIdleTime == 0 {
		cfg.Database.ConnMaxIdleTime = 10
	}
}

// overrideWithEnv 用环境变量覆盖配置，优先级：环境变量 > config.yaml。
// 数字类变量解析失败必须报错，不能被静默忽略。
func overrideWithEnv(cfg *config.Config) error {
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("环境变量 DB_PORT 非法：%q", port)
		}
		cfg.Database.Port = p
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("环境变量 SERVER_PORT 非法：%q", port)
		}
		cfg.Server.Port = p
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
		cfg.AI.Cover.APIKey = key
	}
	if base := os.Getenv("DASHSCOPE_BASE"); base != "" {
		cfg.AI.Cover.BaseURL = base
	}
	if dir := os.Getenv("COVERS_DIR"); dir != "" {
		cfg.AI.Cover.CoversDir = dir
	}
	// COVER_ENABLED=false 显式关闭封面生成（默认由 yaml 的 ai.cover.enabled 决定）
	if v := os.Getenv("COVER_ENABLED"); v != "" {
		cfg.AI.Cover.Enabled = strings.EqualFold(v, "true") || v == "1"
	}
	// 图片存储：local / cos，COS 凭据只用环境变量注入
	if t := os.Getenv("STORAGE_TYPE"); t != "" {
		cfg.Storage.Type = t
	}
	if id := os.Getenv("COS_SECRET_ID"); id != "" {
		cfg.Storage.Cos.SecretID = id
	}
	if key := os.Getenv("COS_SECRET_KEY"); key != "" {
		cfg.Storage.Cos.SecretKey = key
	}
	if bucket := os.Getenv("COS_BUCKET"); bucket != "" {
		cfg.Storage.Cos.Bucket = bucket
	}
	if region := os.Getenv("COS_REGION"); region != "" {
		cfg.Storage.Cos.Region = region
	}
	if domain := os.Getenv("COS_DOMAIN"); domain != "" {
		cfg.Storage.Cos.Domain = domain
	}
	if prefix := os.Getenv("COS_PREFIX"); prefix != "" {
		cfg.Storage.Cos.Prefix = prefix
	}
	return nil
}

// LogConfig 在 logger 就绪后输出降级告警与生效配置摘要，
// 便于确认环境变量是否真的覆盖成功（此前 global.Log 尚不可用）。
func LogConfig(warnings []string) {
	for _, w := range warnings {
		global.Log.Warn("配置告警", zap.String("detail", w))
	}
	coverStatus := "disabled"
	if global.Config.AI.Cover.Enabled {
		coverStatus = "enabled"
	}
	global.Log.Info("配置校验通过",
		zap.Int("port", global.Config.Server.Port),
		zap.String("storage", strings.ToLower(global.Config.Storage.Type)),
		zap.String("ai_cover", coverStatus),
		zap.String("db", fmt.Sprintf("%s:%d/%s",
			global.Config.Database.Host, global.Config.Database.Port, global.Config.Database.DBName)),
	)
}
