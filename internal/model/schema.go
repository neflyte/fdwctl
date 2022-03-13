package model

import (
	"fmt"
	"strings"
)

// Grants represents a permission configuration for an imported remote schema
type Grants struct {
	// Users is the list of users to apply the permissions to
	Users []string `yaml:"users" json:"users"`
}

func (g Grants) String() string {
	return fmt.Sprintf("users: {%s}", strings.Join(g.Users, ","))
}

// Schema represents a foreign schema configuration
type Schema struct {
	// ServerName is the name of the foreign server that will be used to import the remote schema
	ServerName string `yaml:"-" json:"-"`
	// LocalSchema is the name of the local schema that the remote schema will be imported into
	LocalSchema string `yaml:"localschema" json:"localschema"`
	// RemoteSchema is the name of the remote schema to import
	RemoteSchema string `yaml:"remoteschema" json:"remoteschema"`
	// ImportENUMs indicates whether remote ENUM types should be auto-created locally before importing
	ImportENUMs bool `yaml:"importenums" json:"importenums"`
	// ENUMConnection specifies the connection string to the remote database for reading ENUM definitions
	ENUMConnection string `yaml:"enumconnection,omitempty" json:"enumconnection,omitempty"`
	// ENUMSecret configures how to retrieve the optional credential for the ENUMConnection connection string
	ENUMSecret Secret `yaml:"enumsecret,omitempty" json:"enumsecret,omitempty"`
	// SchemaGrants is the permission configuration for this foreign schema
	SchemaGrants Grants `yaml:"grants,omitempty" json:"grants,omitempty"`
}

func (s Schema) String() string {
	return fmt.Sprintf(
		"name: %s, localschema: %s, remoteschema: %s, importenumps: %t, enumconnection: %s, enumsecret: %s, grants: %s",
		s.ServerName,
		s.LocalSchema,
		s.RemoteSchema,
		s.ImportENUMs,
		s.ENUMConnection,
		s.ENUMSecret,
		s.SchemaGrants,
	)
}
