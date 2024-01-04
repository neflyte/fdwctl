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
	ENUMSecret     Secret `yaml:"enumsecret,omitempty" json:"enumsecret,omitempty"`
	ServerName     string `yaml:"-" json:"-"`
	LocalSchema    string `yaml:"localschema" json:"localschema"`
	RemoteSchema   string `yaml:"remoteschema" json:"remoteschema"`
	ENUMConnection string `yaml:"enumconnection,omitempty" json:"enumconnection,omitempty"`
	SchemaGrants   Grants `yaml:"grants,omitempty" json:"grants,omitempty"`
	ImportENUMs    bool   `yaml:"importenums" json:"importenums"`
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

type SchemaEnum struct {
	Schema string
	Name   string
}

func (se *SchemaEnum) String() string {
	return fmt.Sprintf("%s.%s", se.Schema, se.Name)
}
