package config

import (
	"fmt"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	"os"
	"path"
)

const (
	FileName        = "config.yaml"
	XDGAppConfigDir = "fdwctl"
)

var (
	instance AppConfig
)

type AppConfig struct {
	FDWConnection string `yaml:"FDWConnection"`
}

func init() {
	instance = AppConfig{}
}

func Instance() *AppConfig {
	return &instance
}

func Load(ac *AppConfig) error {
	log := logger.Root().
		WithField("function", "Load")
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			log.Errorf("error getting user home directory: %s", err)
		}
		if homedir != "" {
			xdgConfigHome = path.Join(homedir, ".config")
		}
	}
	cfgFullPath := path.Join(xdgConfigHome, XDGAppConfigDir, FileName)
	fs := afero.NewOsFs()
	configFileExists, err := afero.Exists(fs, cfgFullPath)
	if err != nil {
		return fmt.Errorf("error checking existence of config file: %s", err)
	}
	if !configFileExists {
		log.Debugf("config file does not exist")
		return nil
	}
	rawConfigBytes, err := afero.ReadFile(fs, cfgFullPath)
	if err != nil {
		return fmt.Errorf("error reading config file: %s", err)
	}
	err = yaml.Unmarshal(rawConfigBytes, ac)
	if err != nil {
		return fmt.Errorf("error unmarshaling config: %s", err)
	}
	return nil
}
