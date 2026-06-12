package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Config 总配置
type Config struct {
	MySQL      MySql      `toml:"mysql"`
	Redis      Redis      `toml:"redis"`
	Server     Server     `toml:"server"`
	Ratelimit  Ratelimit  `toml:"ratelimit"`
	AsyncFlush AsyncFlush `toml:"asyncflush"`
	URLCheck   URLCheck   `toml:"urlcheck"`
	Cache      Cache      `toml:"cache"`
}

type MySql struct {
	DSN string `toml:"dsn"`
}

type Redis struct {
	Addr     string `toml:"addr"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`
}

type Server struct {
	Port   int    `toml:"port"`
	Domain string `toml:"domain"`
}

type Ratelimit struct {
	EverySeconds int `toml:"every_seconds"`
	Burst        int `toml:"burst"`
}

type AsyncFlush struct {
	IntervalSeconds int `toml:"interval_seconds"`
	ScanCount       int `toml:"scan_count"`
}

type URLCheck struct {
	TimeoutSeconds    int `toml:"timeout_seconds"`
	MaxRetries        int `toml:"max_retries"`
	RetryDelaySeconds int `toml:"retry_delay_seconds"`
}

type Cache struct {
	TTLSeconds int `toml:"ttl_seconds"`
}

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
