package options

// EmailOptions definition
type EmailOptions struct {
	Host       string `json:"host"     mapstructure:"host"`
	Port       int    `json:"port"     mapstructure:"port"`
	Username   string `json:"username" mapstructure:"username"`
	Password   string `json:"password"    mapstructure:"password"`
	From       string `json:"from"        mapstructure:"from"`
	UseSSL     bool   `json:"use_ssl"     mapstructure:"use-ssl"`     // For Implicit TLS (usually port 465)
	SkipVerify bool   `json:"skip_verify" mapstructure:"skip-verify"` // Skip TLS certificate verification
}

// NewEmailOptions creates a new EmailOptions object with default parameters.
func NewEmailOptions() *EmailOptions {
	return &EmailOptions{
		Host:       "smtp.exmail.qq.com",
		Port:       465,
		Username:   "notify@littlelights.ai",
		Password:   "",
		From:       "notify@littlelights.ai",
		UseSSL:     true,
		SkipVerify: false,
	}
}

// Validate checks availability of the options.
func (o *EmailOptions) Validate() []error {
	return []error{}
}

// Sanitize returns a copy of the options with sensitive data masked.
func (o *EmailOptions) Sanitize() *EmailOptions {
	sanitized := *o
	if sanitized.Password != "" {
		sanitized.Password = "******"
	}
	return &sanitized
}
