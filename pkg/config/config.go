package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig   `mapstructure:"server"`
	Cameras []CameraConfig `mapstructure:"cameras"`
}

type ServerConfig struct {
	Address string `mapstructure:"address"`
}

type CameraConfig struct {
	Name   string `mapstructure:"name"`
	Device string `mapstructure:"device"`
	Format string `mapstructure:"format"`
	Width  int    `mapstructure:"width"`
	Height int    `mapstructure:"height"`
}

func init() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/camjam")
	viper.AddConfigPath("$HOME/.camjam")
	viper.AddConfigPath(".")
}

func SetConfigFile(file string) {
	viper.SetConfigFile(file)
}

func ReadConfig() Config {
	var c Config

	viper.ReadInConfig()
	viper.Unmarshal(&c)

	return c
}
