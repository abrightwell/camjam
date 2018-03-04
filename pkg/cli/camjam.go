package cli

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/abrightwell/camjam/pkg/config"
	"github.com/abrightwell/camjam/pkg/server"
)

// command options
var (
	logLevel   string
	configFile string
)

func Main() {
	camjamCmd.SetArgs(os.Args[1:])

	if err := camjamCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed running %q\n", os.Args[1])
		os.Exit(1)
	}
}

var camjamCmd = &cobra.Command{
	Use:   "camjam",
	Short: "",
	RunE:  runCamJam,
}

func init() {
	flags := camjamCmd.Flags()

	flags.StringVarP(&configFile, "config", "", "", "path to config file")
	flags.StringVarP(&logLevel, "log-level", "", "info", "logging level")

	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
}

func runCamJam(cmd *cobra.Command, args []string) error {
	level, _ := log.ParseLevel(logLevel)
	log.SetLevel(level)

	if configFile != "" {
		config.SetConfigFile(configFile)
	}

	c := config.ReadConfig()

	s := server.Server{}

	s.Initialize(c)

	s.Start()

	return nil
}
