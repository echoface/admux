package config

import (
	"time"
)

type ServerConfig struct {
	Host string
	Port int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewDefaultConfig() *ServerConfig {
	return &ServerConfig{
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}