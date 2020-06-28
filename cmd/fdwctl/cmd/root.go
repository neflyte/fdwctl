package cmd

import (
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
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initCommand)
	rootCmd.PersistentFlags().StringVar(&logFormat, "logformat", "text", "log output format")
	rootCmd.PersistentFlags().StringVarP(&connectionString, "connection", "c", "", "database connection string (required)")
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(createCmd)
}

func initCommand() {
	logger.SetFormat(logFormat)
}
