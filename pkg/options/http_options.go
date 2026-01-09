package options

import (
	"fmt"
	"time"
)

// HTTPOptions contains http server specific configuration
type HTTPOptions struct {
	Port            int           `json:"port"            mapstructure:"port"`
	ReadTimeout     time.Duration `json:"readTimeout"     mapstructure:"readTimeout"`
	WriteTimeout    time.Duration `json:"writeTimeout"    mapstructure:"writeTimeout"`
	ShutdownTimeout time.Duration `json:"shutdownTimeout" mapstructure:"shutdownTimeout"`
	TrustedProxies  []string      `json:"trustedProxies"  mapstructure:"trustedProxies"`
}

// NewHTTPOptions create a `zero` value instance.
func NewHTTPOptions() *HTTPOptions {
	return &HTTPOptions{
		Port:            8080,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		ShutdownTimeout: 5 * time.Second,
		TrustedProxies:  []string{"127.0.0.1"},
	}
}

// Validate verifies flags passed to HTTPOptions.
func (o *HTTPOptions) Validate() []error {
	var errs []error
	if o.Port <= 0 || o.Port > 65535 {
		errs = append(errs, fmt.Errorf("http port %d must be between 1 and 65535", o.Port))
	}
	if o.ReadTimeout <= 0 {
		errs = append(errs, fmt.Errorf("readTimeout must be greater than 0"))
	}
	if o.WriteTimeout <= 0 {
		errs = append(errs, fmt.Errorf("writeTimeout must be greater than 0"))
	}
	if o.ShutdownTimeout <= 0 {
		errs = append(errs, fmt.Errorf("shutdownTimeout must be greater than 0"))
	}
	return errs
}
