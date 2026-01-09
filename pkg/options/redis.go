package options

import (
	"fmt"
	"time"
)

// RedisOptions contains Redis-specific configuration
type RedisOptions struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool-size"`
	MinIdleConns int           `mapstructure:"min-idle-conns"`
	DialTimeout  time.Duration `mapstructure:"dial-timeout"`
	ReadTimeout  time.Duration `mapstructure:"read-timeout"`
	WriteTimeout time.Duration `mapstructure:"write-timeout"`
	EnableTLS    bool          `mapstructure:"enable-tls"`
	Protocol     int           `mapstructure:"protocol"`
}

// NewRedisOptions create a `zero` value instance.
func NewRedisOptions() *RedisOptions {
	return &RedisOptions{
		Host:         "127.0.0.1",
		Port:         6379,
		Password:     "",
		DB:           0,
		PoolSize:     100,
		MinIdleConns: 10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// Validate verifies flags passed to RedisOptions.
func (o *RedisOptions) Validate() []error {
	errs := []error{}

	if o.Host == "" {
		errs = append(errs, fmt.Errorf("redis host cannot be empty"))
	}
	if o.Port <= 0 || o.Port > 65535 {
		errs = append(errs, fmt.Errorf("redis port %d must be between 1 and 65535", o.Port))
	}
	if o.PoolSize <= 0 {
		errs = append(errs, fmt.Errorf("redis pool size must be greater than 0"))
	}
	if o.MinIdleConns < 0 {
		errs = append(errs, fmt.Errorf("redis min idle conns cannot be negative"))
	}
	return errs
}

// Sanitize returns a copy of the options with sensitive data masked.
func (o *RedisOptions) Sanitize() *RedisOptions {
	sanitized := *o
	if sanitized.Password != "" {
		sanitized.Password = "******"
	}
	return &sanitized
}
