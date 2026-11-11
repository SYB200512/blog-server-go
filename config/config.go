package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Config 全局配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	AI       AIConfig       `yaml:"ai"`
	Storage  StorageConfig  `yaml:"storage"`
	Log      ZapConfig      `yaml:"log"`
}

type ServerConfig struct {
	Port            int    `yaml:"port"`
	Mode            string `yaml:"mode"`
	RequestTimeout  int    `yaml:"request_timeout"`  // 单请求处理预算（秒）
	ShutdownTimeout int    `yaml:"shutdown_timeout"` // 优雅停机等待时间（秒）

	// P0 新增：HTTP Server 级别的超时与请求头上限
	ReadTimeout    int    `yaml:"read_timeout"`     // 读取整个请求（含 body）的最长时间（秒），要 >= 上传预算 60 秒
	WriteTimeout   int    `yaml:"write_timeout"`    // 写出响应的最长时间（秒）
	IdleTimeout    int    `yaml:"idle_timeout"`     // keep-alive 空闲连接保留时间（秒）
	MaxHeaderBytes int    `yaml:"max_header_bytes"` // 请求头大小上限（字节）
	PprofAddr      string `yaml:"pprof_addr"`
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

	// P0 新增：连接生命周期（分钟）
	ConnMaxLifetime int `yaml:"conn_max_lifetime"`  // 小于 MySQL wait_timeout，避免拿到已被服务端断开的连接
	ConnMaxIdleTime int `yaml:"conn_max_idle_time"` // 高峰期后回收空闲连接
}

// JWTConfig 双 token 方案：access 短期、refresh 长期。
// 两个密钥必须不同，否则 refresh token 可能被当作 access token 接受。
type JWTConfig struct {
	AccessTokenSecret      string `yaml:"access_token_secret"`
	RefreshTokenSecret     string `yaml:"refresh_token_secret"`
	AccessTokenExpiryTime  string `yaml:"access_token_expiry_time"`  // Go duration，如 30m
	RefreshTokenExpiryTime string `yaml:"refresh_token_expiry_time"` // Go duration，如 168h
	Issuer                 string `yaml:"issuer"`
	CookieSecure           bool   `yaml:"cookie_secure"` // 生产 HTTPS 环境必须为 true
}

// RedisConfig 缓存配置。Redis 是硬依赖（必须配置 addr），但运行期故障不致命：
// 缓存层 fail-open + 熔断会降级为直连数据库，Redis 恢复后自动生效。
type RedisConfig struct {
	Addr         string `yaml:"addr"`          // 如 127.0.0.1:6379
	Password     string `yaml:"password"`      // 环境变量 REDIS_PASSWORD 注入
	DB           int    `yaml:"db"`            // 逻辑库索引，默认 1，与其他应用隔离
	PoolSize     int    `yaml:"pool_size"`     // 连接池大小，默认 10
	DialTimeout  int    `yaml:"dial_timeout"`  // 连接建立上限（秒），默认 2
	ReadTimeout  int    `yaml:"read_timeout"`  // 单次读上限（毫秒），默认 300：缓存慢不允许拖长请求
	WriteTimeout int    `yaml:"write_timeout"` // 单次写上限（毫秒），默认 300
	KeyPrefix    string `yaml:"key_prefix"`    // key 前缀，默认 blog
}

// AIConfig AI 相关配置
type AIConfig struct {
	Cover CoverConfig `yaml:"cover"`
}

// CoverConfig AI 封面生成配置。Enabled 决定 cover 功能是否开启：
//   - enabled=true  且 api_key 为空 → 启动硬失败（否则发文后异步生成静默失败）
//   - enabled=false 且 api_key 为空 → 仅 warn，封面功能关闭
type CoverConfig struct {
	Enabled      bool   `yaml:"enabled"`       // 是否启用 AI 封面生成
	APIKey       string `yaml:"api_key"`       // DashScope API Key，环境变量 DASHSCOPE_API_KEY
	BaseURL      string `yaml:"base_url"`      // DashScope 服务地址
	CoversDir    string `yaml:"covers_dir"`    // 本地存储时封面/图片落盘目录，cos 模式不使用
	CoverTimeout int    `yaml:"cover_timeout"` // 异步生成封面超时（秒）
}

// StorageConfig 图片存储配置：type 决定使用本地目录还是对象存储
type StorageConfig struct {
	Type    string    `yaml:"type"`     // local / cos
	MaxSize int       `yaml:"max_size"` // 单张图片大小上限（MB）
	Cos     CosConfig `yaml:"cos"`
}

// CosConfig 腾讯云对象存储配置
type CosConfig struct {
	SecretID  string `yaml:"secret_id"`
	SecretKey string `yaml:"secret_key"`
	Bucket    string `yaml:"bucket"` // 存储桶名称，如 glog-picture-1250000000
	Region    string `yaml:"region"` // 地域，如 ap-guangzhou
	Domain    string `yaml:"domain"` // 自定义访问域名，留空则用默认域名
	Prefix    string `yaml:"prefix"` // 对象键前缀，如 blog/
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

// 存储类型枚举
const (
	StorageLocal = "local"
	StorageCos   = "cos"
)

// placeholderSecrets 示例/占位密钥，防止把 example 文件改个名就上线
var placeholderSecrets = map[string]struct{}{
	"change-me":   {},
	"changeme":    {},
	"secret":      {},
	"your-secret": {},
}

func isPlaceholderSecret(s string) bool {
	_, ok := placeholderSecrets[strings.ToLower(s)]
	return ok
}

// Validate 启动前校验，必须在 applyDefaults 与环境变量覆盖之后调用。
//
// 返回值约定（同时产出 warnings 与 error，不是二选一）：
//   - err 非 nil：硬性缺陷，调用方必须退出；用 errors.Join 聚合全部问题一次性输出，
//     避免运维改一个字段重启一次。
//   - warnings：可选功能降级提示，启动后由调用方记 warn 日志。
//
// 校验同时负责把非法但可恢复的字段归一化（如 log.level 非法时置为 info 并告警），
// 保证后续消费方拿到的 Config 是自洽的。
func (c *Config) Validate() (warnings []string, err error) {
	var problems []error

	// —— 安全相关：为空或占位值一律拒绝 ——
	accessSecret := strings.TrimSpace(c.JWT.AccessTokenSecret)
	refreshSecret := strings.TrimSpace(c.JWT.RefreshTokenSecret)
	for name, item := range map[string]struct{ value, env string }{
		"jwt.access_token_secret":  {accessSecret, "JWT_ACCESS_SECRET"},
		"jwt.refresh_token_secret": {refreshSecret, "JWT_REFRESH_SECRET"},
	} {
		switch {
		case item.value == "":
			problems = append(problems, fmt.Errorf("%s 未配置，可用环境变量 %s 注入", name, item.env))
		case len(item.value) < 32:
			problems = append(problems, fmt.Errorf("%s 长度不足 32 字节（当前 %d），弱密钥可伪造 token", name, len(item.value)))
		case isPlaceholderSecret(item.value):
			problems = append(problems, fmt.Errorf("%s 仍是占位值 %q，必须替换为随机串", name, item.value))
		}
	}
	if accessSecret != "" && accessSecret == refreshSecret {
		problems = append(problems, errors.New("jwt.access_token_secret 与 jwt.refresh_token_secret 不能相同"))
	}
	accessTTL, accessErr := time.ParseDuration(c.JWT.AccessTokenExpiryTime)
	if accessErr != nil {
		problems = append(problems, fmt.Errorf("jwt.access_token_expiry_time 非法：%v（示例 30m）", accessErr))
	}
	refreshTTL, refreshErr := time.ParseDuration(c.JWT.RefreshTokenExpiryTime)
	if refreshErr != nil {
		problems = append(problems, fmt.Errorf("jwt.refresh_token_expiry_time 非法：%v（示例 168h）", refreshErr))
	}
	if accessErr == nil && refreshErr == nil && accessTTL >= refreshTTL {
		problems = append(problems, errors.New("jwt.access_token_expiry_time 必须小于 jwt.refresh_token_expiry_time"))
	}

	// —— 数据库连接要素 ——
	if c.Database.Host == "" {
		problems = append(problems, errors.New("database.host 未配置，可用环境变量 DB_HOST 注入"))
	}
	if c.Database.User == "" {
		problems = append(problems, errors.New("database.user 未配置，可用环境变量 DB_USER 注入"))
	}
	if c.Database.DBName == "" {
		problems = append(problems, errors.New("database.dbname 未配置，可用环境变量 DB_NAME 注入"))
	}
	// 生产环境不允许空密码；开发环境只告警
	if c.Database.Password == "" {
		if strings.ToLower(c.Server.Mode) == "release" {
			problems = append(problems, errors.New("database.password 为空：生产环境(server.mode=release)必须设置 DB_PASSWORD"))
		} else {
			warnings = append(warnings, "database.password 为空，仅建议本地开发环境使用")
		}
	}

	// —— Redis：addr 缺失属于配置缺陷（硬失败）；连接失败不在此校验，由运行期降级兜底 ——
	if strings.TrimSpace(c.Redis.Addr) == "" {
		problems = append(problems, errors.New("redis.addr 未配置，可用环境变量 REDIS_ADDR 注入"))
	}
	if c.Redis.DB < 0 || c.Redis.DB > 15 {
		problems = append(problems, fmt.Errorf("redis.db 非法：%d（Redis 默认 16 个逻辑库，0-15）", c.Redis.DB))
	}
	if c.Redis.Password == "" && strings.ToLower(c.Server.Mode) == "release" {
		warnings = append(warnings, "redis.password 为空：仅建议同机/内网部署使用")
	}

	// —— 存储类型必须合法；cos 必须带全凭据 ——
	switch strings.ToLower(c.Storage.Type) {
	case StorageLocal:
		// 本地存储下 covers 目录承载全部图片落盘，不可写则上传链路全挂 → 硬失败
		if err := ensureDirWritable(c.AI.Cover.CoversDir); err != nil {
			problems = append(problems, fmt.Errorf("storage.type=local 但 covers 目录不可写(%s): %w", c.AI.Cover.CoversDir, err))
		}
	case StorageCos:
		cos := c.Storage.Cos
		if cos.SecretID == "" {
			problems = append(problems, errors.New("storage.type=cos 但 cos.secret_id 为空，可用环境变量 COS_SECRET_ID 注入"))
		}
		if cos.SecretKey == "" {
			problems = append(problems, errors.New("storage.type=cos 但 cos.secret_key 为空，可用环境变量 COS_SECRET_KEY 注入"))
		}
		if cos.Bucket == "" {
			problems = append(problems, errors.New("storage.type=cos 但 cos.bucket 为空，可用环境变量 COS_BUCKET 注入"))
		}
		if cos.Region == "" {
			problems = append(problems, errors.New("storage.type=cos 但 cos.region 为空，可用环境变量 COS_REGION 注入"))
		}
	default:
		problems = append(problems, fmt.Errorf("storage.type 非法：%q，只支持 local / cos", c.Storage.Type))
	}
	// applyDefaults 已兜底为 5，此处为安全网：0 不代表"不限"，而是会被压到 2 MB
	if c.Storage.MaxSize <= 0 {
		problems = append(problems, errors.New("storage.max_size 必须大于 0"))
	}

	// 端口：0 表示让系统随机分配（测试/容器场景合法），仅拒绝负数和越界
	if c.Server.Port < 0 || c.Server.Port > 65535 {
		problems = append(problems, fmt.Errorf("server.port 非法：%d（0 表示随机分配）", c.Server.Port))
	}

	// —— AI 封面：按 Enabled 决定硬失败还是降级 ——
	if c.AI.Cover.Enabled {
		if c.AI.Cover.APIKey == "" {
			problems = append(problems, errors.New("ai.cover.enabled=true 但 api_key 为空：用环境变量 DASHSCOPE_API_KEY 注入，或将 enabled 设为 false 关闭封面功能"))
		}
	} else if c.AI.Cover.APIKey == "" {
		warnings = append(warnings, "ai.cover.enabled=false 且未配置 DASHSCOPE_API_KEY，AI 封面生成已关闭")
	}

	// —— log.level 非法则归一为 info 并告警（Validate 是唯一归一化点，logger 直接信任该值）——
	switch strings.ToLower(c.Log.Level) {
	case "debug", "info", "warn", "error":
	default:
		warnings = append(warnings, fmt.Sprintf("log.level 非法：%q，已归一为 info", c.Log.Level))
		c.Log.Level = "info"
	}

	return warnings, errors.Join(problems...)
}

// ensureDirWritable 校验目录可创建且可写：创建目录后写一个临时文件再删除。
func ensureDirWritable(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".writetest-*")
	if err != nil {
		return err
	}
	name := f.Name()
	_ = f.Close()
	return os.Remove(name)
}
