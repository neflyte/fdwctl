package cmd

import (
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:     "fdwcli",
		Short:   "A management CLI for PostgreSQL postgres_fdw",
		PreRunE: preDo,
	}
	connectionString = ""
	logFormat        = ""
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initCommand)
	rootCmd.PersistentFlags().StringVar(&logFormat, "logformat", "text", "log output format")
	rootCmd.PersistentFlags().StringVarP(&connectionString, "connection", "c", "", "database connection string")
	// Add commands here
	rootCmd.AddCommand(listCmd)
}

func initCommand() {
	logger.SetFormat(logFormat)
}

func preDo(cmd *cobra.Command, args []string) error {
	return nil
}
