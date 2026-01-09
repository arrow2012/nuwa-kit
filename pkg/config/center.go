package config

import (
	"errors"
	"time"

	"github.com/spf13/viper"
)

var (
	ErrConfigNotFound = errors.New("configuration not found")
	ErrInvalidConfig  = errors.New("invalid configuration")
)

// ConfigCenter defines the interface for configuration center
type ConfigCenter interface {
	GetKey(key string) any
	Close() error
	GetClient() *viper.Viper
}

// ConfigCenterType represents the type of configuration center
type ConfigCenterType string

const (
	ConfigCenterTypeConsul ConfigCenterType = "consul"
	ConfigCenterTypeNacos  ConfigCenterType = "nacos"
	ConfigCenterTypeLocal  ConfigCenterType = "local"
)

// ConfigCenterOptions contains options for configuration center
type ConfigCenterOptions struct {
	Type      ConfigCenterType
	Address   string
	Namespace string
	Timeout   time.Duration
	// Consul specific options
	Token string
	// Nacos specific options
	Group       string
	Username    string
	Password    string
	DataID      string
	Path        string
	ContentType string
}

func NewConfigCenterOptions(t ConfigCenterType, addr string, Path string, contentType string) *ConfigCenterOptions {
	return &ConfigCenterOptions{
		Type:        t,
		Address:     addr,
		Timeout:     5 * time.Second,
		Path:        Path,
		ContentType: contentType,
	}
}

// NewConfigCenter creates a new configuration center based on the type
func NewConfigCenter(opts *ConfigCenterOptions) (ConfigCenter, error) {
	if opts.Timeout == 0 {
		opts.Timeout = 5 * time.Second
	}
	switch opts.Type {
	case ConfigCenterTypeConsul:
		return NewConsulConfigCenter(opts)
	case ConfigCenterTypeLocal:
		return NewLocalConfigCenter(opts)
	default:
		return nil, errors.New("unsupported configuration center type: " + string(opts.Type))
	}
}
