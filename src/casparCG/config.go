package casparcg

import (
	"fmt"
	"net"
)

type Config struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if net.ParseIP(c.Host) == nil {
		return fmt.Errorf("invalid host: %s", c.Host)
	}

	if c.Port == 0 {
		return fmt.Errorf("port is required")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	return nil
}

func (c *Config) Default() {
	def := Config{
		Host: "127.0.0.1",
		Port: 5250,
	}
	*c = def
}
