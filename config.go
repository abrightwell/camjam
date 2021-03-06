package main

import (
	"time"
)

type Config struct {
	Server  ServerConfig   `mapstructure:"server"`
	Cameras []CameraConfig `mapstructure:"cameras"`
}

type ServerConfig struct {
	Address  string        `mapstructure:"address"`
	Interval time.Duration `mapstructure:"interval"`
}

type CameraConfig struct {
	Name   string `mapstructure:"name"`
	Device string `mapstructure:"device"`
	Format string `mapstructure:"format"`
	Width  uint32 `mapstructure:"width"`
	Height uint32 `mapstructure:"height"`
}
