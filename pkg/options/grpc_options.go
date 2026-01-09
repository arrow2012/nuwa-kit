package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

// GRPCOptions defines configuration for gRPC server.
type GRPCOptions struct {
	Port                  int           `json:"port,omitempty"                     mapstructure:"port"`
	MaxConnectionAge      time.Duration `json:"max-connection-age,omitempty"       mapstructure:"max-connection-age"`
	MaxConnectionAgeGrace time.Duration `json:"max-connection-age-grace,omitempty" mapstructure:"max-connection-age-grace"`
	KeepAliveTime         time.Duration `json:"keep-alive-time,omitempty"          mapstructure:"keep-alive-time"`
	KeepAliveTimeout      time.Duration `json:"keep-alive-timeout,omitempty"       mapstructure:"keep-alive-timeout"`
}

// NewGRPCOptions creates a new GRPCOptions with default values.
func NewGRPCOptions() *GRPCOptions {
	return &GRPCOptions{
		Port:                  9090,
		MaxConnectionAge:      time.Duration(time.Hour * 24), // Recycle connections every 24h
		MaxConnectionAgeGrace: time.Duration(time.Minute * 5),
		KeepAliveTime:         time.Duration(time.Minute * 1), // Ping every 1m
		KeepAliveTimeout:      time.Duration(time.Second * 20),
	}
}

// AddFlags adds flags for GRPCOptions to the specified FlagSet.
func (o *GRPCOptions) AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&o.Port, "grpc.port", o.Port, "The port for the gRPC server to listen on.")
	fs.DurationVar(&o.MaxConnectionAge, "grpc.max-connection-age", o.MaxConnectionAge, "A duration for the maximum amount of time a connection may exist before it will be closed.")
	fs.DurationVar(&o.MaxConnectionAgeGrace, "grpc.max-connection-age-grace", o.MaxConnectionAgeGrace, "An additive period after max-connection-age after which the connection will be forcibly closed.")
	fs.DurationVar(&o.KeepAliveTime, "grpc.keep-alive-time", o.KeepAliveTime, "After a duration of this time if the client doesn't see any activity it pings the server to see if the transport is still alive.")
	fs.DurationVar(&o.KeepAliveTimeout, "grpc.keep-alive-timeout", o.KeepAliveTimeout, "After having pinged for keepalive check, the client waits for a duration of this time and if no activity is seen even after that the connection is closed.")
}

// Validate checks GRPCOptions for validation errors.
func (o *GRPCOptions) Validate() []error {
	var errs []error
	if o.Port <= 0 || o.Port > 65535 {
		errs = append(errs, fmt.Errorf("grpc port %d must be between 1 and 65535", o.Port))
	}
	return errs
}
