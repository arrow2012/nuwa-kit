// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package options

import (
	"fmt"
	"runtime"
	"time"
)

// MySQLOptions defines options for mysql database.
type MySQLOptions struct {
	Host                  string        `json:"host,omitempty"                     mapstructure:"host"`
	Username              string        `json:"username,omitempty"                 mapstructure:"username"`
	Password              string        `json:"-"                                  mapstructure:"password"`
	Database              string        `json:"database"                           mapstructure:"database"`
	MaxIdleConnections    int           `json:"max-idle-connections,omitempty"     mapstructure:"max-idle-connections"`
	MaxOpenConnections    int           `json:"max-open-connections,omitempty"     mapstructure:"max-open-connections"`
	MaxConnectionLifeTime time.Duration `json:"max-connection-life-time,omitempty" mapstructure:"max-connection-life-time"`
	LogLevel              int           `json:"log-level"                          mapstructure:"log-level"`
	AutoMigrate           bool          `json:"auto-migrate"                       mapstructure:"auto-migrate"`
}

// NewMySQLOptions create a `zero` value instance.
func NewMySQLOptions() *MySQLOptions {
	cpuCores := runtime.NumCPU()
	maxOpen := cpuCores * 20
	maxIdle := cpuCores * 10
	if maxOpen < 10 {
		maxOpen = 10
	}
	if maxIdle < 5 {
		maxIdle = 5
	}

	return &MySQLOptions{
		Host:                  "127.0.0.1:3306",
		Username:              "",
		Password:              "",
		Database:              "",
		MaxIdleConnections:    maxIdle,
		MaxOpenConnections:    maxOpen,
		MaxConnectionLifeTime: 30 * time.Minute,
		LogLevel:              1,    // Info
		AutoMigrate:           true, // Default to true for dev convenience
	}
}

// Validate verifies flags passed to MySQLOptions.
func (o *MySQLOptions) Validate() []error {
	errs := []error{}

	if o.MaxIdleConnections <= 0 {
		errs = append(errs, fmt.Errorf("max-idle-connections must be greater than 0"))
	}
	if o.MaxOpenConnections <= 0 {
		errs = append(errs, fmt.Errorf("max-open-connections must be greater than 0"))
	}
	if o.MaxConnectionLifeTime <= 0 {
		errs = append(errs, fmt.Errorf("max-connection-life-time must be greater than 0"))
	}
	return errs
}

// Sanitize returns a copy of the options with sensitive data masked.
func (o *MySQLOptions) Sanitize() *MySQLOptions {
	sanitized := *o
	if sanitized.Password != "" {
		sanitized.Password = "******"
	}
	return &sanitized
}
