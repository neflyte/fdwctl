package cmd

import (
	"fmt"
	"github.com/neflyte/fdwctl/internal/config"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/neflyte/fdwctl/internal/util"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "fdwctl",
		Short: "A management CLI for PostgreSQL postgres_fdw Foreign Data Wrapper",
	}
	logFormat        string
	logLevel         string
	noLogo           bool
	AppVersion       string
	connectionString string
	configFile       string
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initCommand, initConfig)
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "location of program configuration file")
	rootCmd.PersistentFlags().StringVar(&logFormat, "logformat", logger.TextFormat, "log output format [text, json, elastic]")
	rootCmd.PersistentFlags().StringVar(&logLevel, "loglevel", logger.TraceLevel, "log message level [trace, debug, info, warn, error, fatal, panic]")
	rootCmd.PersistentFlags().StringVar(&connectionString, "connection", "", "database connection string")
	rootCmd.PersistentFlags().BoolVar(&noLogo, "nologo", false, "suppress program name and version message")
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(dropCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(desiredStateCmd)
}

func initCommand() {
	logger.SetFormat(logFormat)
	logger.SetLevel(logLevel)
	if !noLogo {
		appVer := "dev"
		if AppVersion != "" {
			appVer = AppVersion
		}
		fmt.Printf("fdwctl v%s\n", appVer)
	}
}

func initConfig() {
	var err error

	log := logger.Root().
		WithField("function", "initConfig")
	if configFile == "" {
		configFile = config.UserConfigFile()
	}
	log.Debugf("configFile: %s", configFile)
	err = config.Load(config.Instance(), configFile)
	if err != nil {
		log.Errorf("error initializing config: %s", err)
		return
	}
	connString := util.StringCoalesce(connectionString, config.Instance().FDWConnection)
	log.Tracef("connString: %s", connString)
	if connString == "" {
		log.Fatal("database connection string is required")
	}
	config.Instance().FDWConnection = connString
}
