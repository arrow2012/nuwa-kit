package options

// ConfigCenterOptions contains configuration center options
type ConfigCenterOptions struct {
	Type      string `json:"type" mapstructure:"type"` // nacos, consul, etc.
	Address   string `json:"address" mapstructure:"address"`
	Namespace string `json:"namespace" mapstructure:"namespace"`
}

// NewServerOptions create a `zero` value instance.
func NewConfigCenterOptions() *ConfigCenterOptions {
	return &ConfigCenterOptions{
		Type:      "nacos",
		Address:   "127.0.0.1:8848",
		Namespace: "public",
	}
}
