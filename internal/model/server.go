package model

import (
	"fmt"
	"strings"
)

// ForeignServer represents a Postgres foreign server including related user mappings and remote schemas
type ForeignServer struct {
	// Name is the display name of the foreign server
	Name string `yaml:"name" json:"name"`
	// Host is the hostname of the foreign server
	Host string `yaml:"host" json:"host"`
	// Port is the port number of the foreign server
	Port int `yaml:"port" json:"port"`
	// DB is the name of the database on the foreign server
	DB string `yaml:"db" json:"db"`
	// Wrapper is the name of the foreign data wrapper to use with this foreign server
	Wrapper string `yaml:"wrapper,omitempty" json:"wrapper,omitempty"`
	// Owner is the owner of the foreign server
	Owner string `yaml:"-" json:"-"`
	// UserMaps is the list of user mappings to create using this foreign server
	UserMaps []UserMap `yaml:"UserMap,omitempty" json:"UserMap,omitempty"`
	// Schemas is the list of remote schemas to import from this foreign server
	Schemas []Schema `yaml:"Schemas,omitempty" json:"Schemas,omitempty"`
}

// Equals determines if this object is equal to the supplied object
func (fs *ForeignServer) Equals(fserver ForeignServer) bool {
	return fs.Name == fserver.Name && fs.Host == fserver.Host && fs.Port == fserver.Port &&
		fs.DB == fserver.DB && fs.Wrapper == fserver.Wrapper
}

func (fs *ForeignServer) String() string {
	userMaps := make([]string, len(fs.UserMaps))
	for idx, usermap := range fs.UserMaps {
		userMaps[idx] = usermap.LocalUser
	}
	schemas := make([]string, len(fs.Schemas))
	for idx, schema := range fs.Schemas {
		schemas[idx] = schema.LocalSchema
	}
	return fmt.Sprintf(
		"name: %s, host: %s, port: %d, db: %s, wrapper: %s, owner: %s, usermaps: {%s}, schemas: {%s}",
		fs.Name,
		fs.Host,
		fs.Port,
		fs.DB,
		fs.Wrapper,
		fs.Owner,
		strings.Join(userMaps, ","),
		strings.Join(schemas, ","),
	)
}
