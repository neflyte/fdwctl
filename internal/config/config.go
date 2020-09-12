/*
Package config contains data structures and methods for handling application configuration
*/
package config

import (
	"encoding/json"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/neflyte/fdwctl/internal/model"
	"github.com/neflyte/fdwctl/internal/util"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"strings"
)

// appConfig is the application configuration structure. Configuration files are unmarshaled into it directly.
type appConfig struct {
	// FDWConnection is the Postgres database connection string for the FDW database
	FDWConnection string `yaml:"FDWConnection" json:"FDWConnection"`
	// FDWConnectionSecret configures how to retrieve the optional credential for the FDWConnection connection string
	FDWConnectionSecret model.Secret `yaml:"FDWConnectionSecret,omitempty" json:"FDWConnectionSecret,omitempty"`
	// DesiredState defines the desired state configuration of the FDW database
	DesiredState model.DesiredState `yaml:"DesiredState,omitempty" json:"DesiredState,omitempty"`
	// dbConnectionString is the calculated connection string for the FDW database
	dbConnectionString string
}

const (
	// fileName is the default configuration file name
	fileName = "config.yaml"
	// xdgAppConfigDir is the subdirectory of XDG_CONFIG_HOME that the configuration file lives in
	xdgAppConfigDir = "fdwctl"
)

var (
	// configInstance is the singleton instance of the application configuration
	configInstance *appConfig
)

// Instance returns the singleton instance of the application configuration
func Instance() *appConfig {
	if configInstance == nil {
		configInstance = &appConfig{
			FDWConnectionSecret: model.Secret{},
			DesiredState: model.DesiredState{
				Extensions: make([]model.Extension, 0),
				Servers:    make([]model.ForeignServer, 0),
			},
		}
	}
	return configInstance
}

// UserConfigFile returns the resolved path and filename of the application configuration file
func UserConfigFile() string {
	log := logger.Log().
		WithField("function", "UserConfigFile")
	xdgConfigHome, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" || !ok {
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
	return path.Join(xdgConfigHome, xdgAppConfigDir, fileName)
}

// Load reads the specified file into the application configuration struct
func Load(ac *appConfig, fileName string) error {
	log := logger.Log().
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
	log.Tracef("loaded config: %#v", ac)
	return nil
}

// GetDatabaseConnectionString returns the calculated connection string for the FDW database
func (ac *appConfig) GetDatabaseConnectionString() string {
	if ac.dbConnectionString == "" {
		ac.dbConnectionString = util.ResolveConnectionString(ac.FDWConnection, &ac.FDWConnectionSecret)
	}
	return ac.dbConnectionString
}
