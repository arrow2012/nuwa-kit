package options

import "fmt"

// WeComOptions contains WeCom-specific configuration
type WeComOptions struct {
	CorpID      string `json:"corpId" mapstructure:"corpId"`
	AgentID     string `json:"agentId" mapstructure:"agentId"`
	Secret      string `json:"secret" mapstructure:"secret"`
	RedirectURI string `json:"redirectUri" mapstructure:"redirectUri"`
}

// NewWeComOptions create a `zero` value instance.
func NewWeComOptions() *WeComOptions {
	return &WeComOptions{
		CorpID:      "",
		AgentID:     "",
		Secret:      "",
		RedirectURI: "",
	}
}

// Validate verifies flags passed to WeComOptions.
func (o *WeComOptions) Validate() []error {
	errs := []error{}
	if o.CorpID == "" {
		errs = append(errs, fmt.Errorf("wecom corpId cannot be empty"))
	}
	if o.AgentID == "" {
		errs = append(errs, fmt.Errorf("wecom agentId cannot be empty"))
	}
	if o.Secret == "" {
		errs = append(errs, fmt.Errorf("wecom secret cannot be empty"))
	}
	return errs
}

// Sanitize returns a copy of the options with sensitive data masked.
func (o *WeComOptions) Sanitize() *WeComOptions {
	sanitized := *o
	if sanitized.Secret != "" {
		sanitized.Secret = "******"
	}
	return &sanitized
}
