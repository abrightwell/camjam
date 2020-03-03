package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	config "github.com/spf13/viper"
)

// command options
var (
	logLevel   string
	configFile string
)

func init() {
	flag.StringVarP(&configFile, "config", "", "", "path to config file")
	flag.StringVarP(&logLevel, "log-level", "", "info", "logging level")

	config.SetConfigType("yaml")
	config.SetConfigName("config")
	config.AddConfigPath("/etc/camjam")
	config.AddConfigPath("$HOME/.camjam")
	config.AddConfigPath(".")

	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	log.SetOutput(os.Stdout)
}

func main() {
	flag.Parse()

	level, _ := log.ParseLevel(logLevel)
	log.SetLevel(level)

	if configFile != "" {
		config.SetConfigFile(configFile)
	}

	var c Config

	if err := config.ReadInConfig(); err != nil {
		log.Fatalf("could not read in config: %s", err.Error())
	}

	if err := config.Unmarshal(&c); err != nil {
		log.Fatalf("could not unmarshal config: %s", err.Error())
	}

	s := Server{}

	s.Initialize(c)

	s.Start()
}
