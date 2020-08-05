package config

import (
	"encoding/json"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/neflyte/fdwctl/internal/model"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"strings"
)

const (
	FileName        = "config.yaml"
	XDGAppConfigDir = "fdwctl"
)

var (
	instance AppConfig
)

type DesiredState struct {
	Extensions []model.Extension     `yaml:"Extensions,omitempty" json:"Extensions,omitempty"`
	Servers    []model.ForeignServer `yaml:"Servers,omitempty" json:"Servers,omitempty"`
}

type AppConfig struct {
	FDWConnection string       `yaml:"FDWConnection"`
	DesiredState  DesiredState `yaml:"DesiredState,omitempty"`
}

func init() {
	instance = AppConfig{
		DesiredState: DesiredState{
			Extensions: make([]model.Extension, 0),
			Servers:    make([]model.ForeignServer, 0),
		},
	}
}

func Instance() *AppConfig {
	return &instance
}

func UserConfigFile() string {
	log := logger.Root().
		WithField("function", "UserConfigFile")
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			log.Errorf("error getting user home directory: %s", err)
		}
		if homedir != "" {
			xdgConfigHome = path.Join(homedir, ".config")
		} else {
			xdgConfigHome = "."
		}
	}
	return path.Join(xdgConfigHome, XDGAppConfigDir, FileName)
}

func Load(ac *AppConfig, fileName string) error {
	log := logger.Root().
		WithField("function", "Load")
	fs := afero.NewOsFs()
	configFileExists, err := afero.Exists(fs, fileName)
	if err != nil {
		return logger.ErrorfAsError(log, "error checking existence of config file: %s", err)
	}
	if !configFileExists {
		log.Debugf("config file does not exist")
		return nil
	}
	rawConfigBytes, err := afero.ReadFile(fs, fileName)
	if err != nil {
		return logger.ErrorfAsError(log, "error reading config file: %s", err)
	}
	if strings.HasSuffix(fileName, "json") {
		err = json.Unmarshal(rawConfigBytes, ac)
	} else {
		err = yaml.Unmarshal(rawConfigBytes, ac)
	}
	if err != nil {
		return logger.ErrorfAsError(log, "error unmarshaling config: %s", err)
	}
	return nil
}
