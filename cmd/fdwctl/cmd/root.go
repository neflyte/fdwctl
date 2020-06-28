package cmd

import (
	"fmt"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "fdwctl",
		Short: "A management CLI for PostgreSQL postgres_fdw",
	}
	connectionString string
	logFormat        string
	logLevel         string
	noLogo           bool
	AppVersion       string
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initCommand)
	rootCmd.Flags().StringVar(&logFormat, "logformat", logger.TextFormat, "log output format [text, json]")
	rootCmd.Flags().StringVar(&logLevel, "loglevel", logger.TraceLevel, "log message level [trace, debug, info, warn, error, fatal, panic]")
	rootCmd.PersistentFlags().StringVarP(&connectionString, "connection", "c", "", "database connection string")
	_ = rootCmd.MarkFlagRequired("connection")
	rootCmd.Flags().BoolVar(&noLogo, "nologo", false, "suppress program name and version message")
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(dropCmd)
	rootCmd.AddCommand(editCmd)
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
