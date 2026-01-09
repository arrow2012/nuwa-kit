package config

import (
	"github.com/spf13/viper"
)

// LocalConfigCenter implements ConfigCenter for local testing
type LocalConfigCenter struct {
	opts *ConfigCenterOptions
	v    *viper.Viper
}

func NewLocalConfigCenter(opts *ConfigCenterOptions) (*LocalConfigCenter, error) {
	v := viper.New()
	// NOTE: We do NOT set specific IAM defaults here.
	// The caller should use GetClient() to set defaults if needed, or pass them in options in future.

	return &LocalConfigCenter{opts: opts, v: v}, nil
}

func (c *LocalConfigCenter) GetKey(key string) any {
	return c.v.Get(key)
}

func (c *LocalConfigCenter) Close() error {
	return nil
}

func (c *LocalConfigCenter) GetClient() *viper.Viper {
	return c.v
}
