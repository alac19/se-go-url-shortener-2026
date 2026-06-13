// Package config provides configuration management for short link services.
package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Config 总配置。
type Config struct {
	MySQL      MySql      `toml:"mysql"`
	Redis      Redis      `toml:"redis"`
	Server     Server     `toml:"server"`
	Ratelimit  Ratelimit  `toml:"ratelimit"`
	AsyncFlush AsyncFlush `toml:"asyncflush"`
	URLCheck   URLCheck   `toml:"urlcheck"`
	Cache      Cache      `toml:"cache"`
	Log        Log        `toml:"log"`
}

// MySql 定义 MySQL 数据库连接配置。
type MySql struct {
	DSN string `toml:"dsn"` // 数据源名称（Data Source Name）
}

// Redis 定义 Redis 缓存配置。
type Redis struct {
	Addr     string `toml:"addr"`     // 服务器地址
	Password string `toml:"password"` // 密码
	DB       int    `toml:"db"`       // 数据库编号
}

// Server 定义 HTTP 服务器配置。
type Server struct {
	Port   int    `toml:"port"`   // 监听端口
	Domain string `toml:"domain"` // 短链接域名（包含协议和端口，例如 http://localhost:8080/）
}

// Ratelimit 定义限流器配置。
type Ratelimit struct {
	EverySeconds int `toml:"every_seconds"` // 生成令牌的时间间隔（秒）
	Burst        int `toml:"burst"`         // 令牌桶容量（突发请求数）
}

// AsyncFlush 定义异步统计写入配置。
type AsyncFlush struct {
	IntervalSeconds int `toml:"interval_seconds"` // 刷新间隔（秒）
	ScanCount       int `toml:"scan_count"`       // SCAN 命令每次返回的键数量提示
}

// URLCheck 定义 URL 可达性检查配置。
type URLCheck struct {
	TimeoutSeconds    int `toml:"timeout_seconds"`     // HTTP 请求超时（秒）
	MaxRetries        int `toml:"max_retries"`         // 最大重试次数
	RetryDelaySeconds int `toml:"retry_delay_seconds"` // 重试间隔（秒）
}

// Cache 定义缓存配置。
type Cache struct {
	TTLSeconds int `toml:"ttl_seconds"` // 缓存过期时间（秒）
}

// Log 定义日志配置。
type Log struct {
	Level    string `toml:"level"`     // 日志级别（debug, info, warn, error）
	FilePath string `toml:"file_path"` // 日志文件路径，为空则只输出到控制台
}

// LoadConfig 从指定路径加载 TOML 配置文件，并校验配置合法性。
func LoadConfig(path string) (*Config, error) {
	var config Config

	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, fmt.Errorf("decode config failed: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

// Validate 校验配置项的合法性。
func (c *Config) Validate() error {
	// 服务器配置
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port 必须在 1-65535 之间，当前值: %d", c.Server.Port)
	}
	if c.Server.Domain == "" {
		return fmt.Errorf("server.domain 不能为空")
	}

	// 限流配置
	if c.Ratelimit.EverySeconds <= 0 {
		return fmt.Errorf("ratelimit.every_seconds 必须 > 0，当前值: %f", c.Ratelimit.EverySeconds)
	}
	if c.Ratelimit.Burst <= 0 {
		return fmt.Errorf("ratelimit.burst 必须 > 0，当前值: %d", c.Ratelimit.Burst)
	}

	// 异步写入配置
	if c.AsyncFlush.IntervalSeconds <= 0 {
		return fmt.Errorf("asyncflush.interval_seconds 必须 > 0，当前值: %d", c.AsyncFlush.IntervalSeconds)
	}
	if c.AsyncFlush.ScanCount <= 0 {
		return fmt.Errorf("asyncflush.scan_count 必须 > 0，当前值: %d", c.AsyncFlush.ScanCount)
	}

	// URL 可达性检查配置
	if c.URLCheck.TimeoutSeconds <= 0 {
		return fmt.Errorf("urlcheck.timeout_seconds 必须 > 0，当前值: %d", c.URLCheck.TimeoutSeconds)
	}
	if c.URLCheck.MaxRetries < 0 {
		return fmt.Errorf("urlcheck.max_retries 不能为负数，当前值: %d", c.URLCheck.MaxRetries)
	}
	if c.URLCheck.MaxRetries > 0 && c.URLCheck.RetryDelaySeconds <= 0 {
		return fmt.Errorf("urlcheck.retry_delay_seconds 在 max_retries > 0 时必须 > 0，当前值: %d", c.URLCheck.RetryDelaySeconds)
	}

	// 缓存配置
	if c.Cache.TTLSeconds <= 0 {
		return fmt.Errorf("cache.ttl_seconds 必须 > 0，当前值: %d", c.Cache.TTLSeconds)
	}

	// MySQL 配置
	if c.MySQL.DSN == "" {
		return fmt.Errorf("mysql.dsn 不能为空")
	}

	// Redis 配置
	if c.Redis.Addr == "" {
		return fmt.Errorf("redis.addr 不能为空")
	}
	if c.Redis.DB < 0 {
		return fmt.Errorf("redis.db 不能为负数，当前值: %d", c.Redis.DB)
	}

	return nil
}
