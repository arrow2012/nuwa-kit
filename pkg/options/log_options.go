package options

import "github.com/spf13/pflag"

// LogOptions contains configuration for logging.
type LogOptions struct {
	Level       string   `json:"level" mapstructure:"level"`
	Format      string   `json:"format" mapstructure:"format"`
	OutputPaths []string `json:"output-paths" mapstructure:"output-paths"`
	EnableSQL   bool     `json:"enable-sql" mapstructure:"enable-sql"`

	// Rotation Config
	EnableRotation bool   `json:"enable-rotation" mapstructure:"enable-rotation"`
	RotateLogPath  string `json:"rotate-log-path" mapstructure:"rotate-log-path"` // e.g. /var/log/iam/iam.log
	MaxSize        int    `json:"max-size" mapstructure:"max-size"`               // Megabytes
	MaxBackups     int    `json:"max-backups" mapstructure:"max-backups"`
	MaxAge         int    `json:"max-age" mapstructure:"max-age"` // Days
	Compress       bool   `json:"compress" mapstructure:"compress"`
}

// NewLogOptions creates a new LogOptions object with default parameters.
func NewLogOptions() *LogOptions {
	return &LogOptions{
		Level:       "info",
		Format:      "console",
		OutputPaths: []string{"stdout"},
		EnableSQL:   false,

		EnableRotation: false,
		RotateLogPath:  "log/iam.log",
		MaxSize:        100, // 100MB
		MaxBackups:     3,
		MaxAge:         7, // 7 Days
		Compress:       true,
	}
}

// Validate checks if the options are valid.
func (o *LogOptions) Validate() []error {
	return nil
}

// AddFlags adds flags for the log options to the specified FlagSet.
func (o *LogOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Level, "log.level", o.Level, "Log level (debug, info, warn, error, fatal, panic)")
	fs.StringVar(&o.Format, "log.format", o.Format, "Log format (json, console)")
	fs.StringSliceVar(&o.OutputPaths, "log.output-paths", o.OutputPaths, "Log output paths")
	fs.BoolVar(&o.EnableSQL, "log.enable-sql", o.EnableSQL, "Enable SQL logging")
}

// Complete completes all the required options.
func (o *LogOptions) Complete() error {
	return nil
}
