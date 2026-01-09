package options

import "fmt"

// RateLimitOptions contains rate limiting configuration
type RateLimitOptions struct {
	// 全局默认限流配置
	DefaultRate  float64 `json:"defaultRate" mapstructure:"defaultRate"`   // 每秒请求数
	DefaultBurst int     `json:"defaultBurst" mapstructure:"defaultBurst"` // 桶容量

	// 是否启用限流
	Enabled bool `json:"enabled" mapstructure:"enabled"`

	// 是否启用管理API
	EnableAdminAPI bool `json:"enableAdminAPI" mapstructure:"enableAdminAPI"`

	// 管理API路径前缀
	AdminAPIPrefix string `json:"adminAPIPrefix" mapstructure:"adminAPIPrefix"`

	// 限流器清理间隔（分钟）
	CleanupInterval int `json:"cleanupInterval" mapstructure:"cleanupInterval"`
}

// NewRateLimitOptions create a `zero` value instance.
func NewRateLimitOptions() *RateLimitOptions {
	return &RateLimitOptions{
		DefaultRate:     100.0, // 默认每秒100个请求
		DefaultBurst:    20,    // 默认桶容量20
		Enabled:         true,  // 默认启用
		EnableAdminAPI:  true,  // 默认启用管理API
		AdminAPIPrefix:  "/admin/rate-limit",
		CleanupInterval: 30, // 30分钟清理一次
	}
}

// Validate verifies flags passed to RateLimitOptions.
func (o *RateLimitOptions) Validate() []error {
	errs := []error{}

	if o.DefaultRate < 0 {
		errs = append(errs, fmt.Errorf("defaultRate cannot be negative"))
	}

	if o.DefaultBurst < 0 {
		errs = append(errs, fmt.Errorf("defaultBurst cannot be negative"))
	}

	if o.CleanupInterval < 1 {
		errs = append(errs, fmt.Errorf("cleanupInterval must be at least 1 minute"))
	}

	return errs
}
