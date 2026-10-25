package config

import "fmt"

// Config 全局配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	AI       AIConfig       `yaml:"ai"`
	Log      ZapConfig      `yaml:"log"`
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

// ZapConfig zap 日志配置（参考 go-blog 项目）
type ZapConfig struct {
	Level      string `yaml:"level"`       // 日志级别: debug / info / warn / error
	Director   string `yaml:"director"`    // 日志输出目录
	MaxSize    int    `yaml:"max_size"`    // 单个日志文件最大大小（MB），超过则切割
	MaxBackups int    `yaml:"max_backups"` // 最多保留的旧日志文件数量
	MaxAge     int    `yaml:"max_age"`     // 日志文件最长保留天数
}

// DSN 返回 MySQL 连接串
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.Charset)
}
