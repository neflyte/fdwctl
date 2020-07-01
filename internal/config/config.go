package config

import (
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/spf13/viper"
	"os"
	"path"
)

var (
	initialized = false
)

func InitConfig() {
	log := logger.Root().
		WithField("function", "InitConfig")
	if !initialized {
		SetDefaults()
		viper.SetConfigName("config.yaml")
		viper.AddConfigPath("$HOME/.config/fdwctl")
		// Ensure config directory exists
		homedir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("error reading user homedir: %s", err)
		}
		configPath := path.Join(homedir, ".config", "fdwctl")
		err = os.MkdirAll(configPath, 0750)
		if err != nil {
			log.Fatalf("error creating config directory: %s", err)
		}
		configFile := path.Join(configPath, "config.yaml")
		err = viper.ReadInConfig()
		if err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				cfile, err := os.Create(configFile)
				if err != nil {
					log.Errorf("error creating new config file: %s", err)
					return
				}
				err = cfile.Close()
				if err != nil {
					log.Errorf("error closing config file: %s", err)
				}
				writeErr := viper.WriteConfig()
				if writeErr != nil {
					log.Errorf("error writing default config: %s", writeErr)
					return
				}
			} else {
				log.Errorf("error reading config: %s", err)
				return
			}
		}
		initialized = true
	}
}

func SetDefaults() {
	viper.SetDefault("FDWConnection", "")
}
