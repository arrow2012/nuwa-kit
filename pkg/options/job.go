package options

// AuthOptions contains authentication-specific configuration
type JobOptions struct {
	Enabled   bool `json:"enabled" mapstructure:"enabled"` // Global switch for Cron service
	CommonJob bool `json:"commonJob" mapstructure:"commonJob"`
}

// NewServerOptions create a `zero` value instance.
func NewJobOptions() *JobOptions {
	return &JobOptions{
		Enabled:   true, // Enable Cron by default
		CommonJob: false,
	}
}

// Validate verifies flags passed to JobOptions.
func (o *JobOptions) Validate() []error {
	return []error{}
}
