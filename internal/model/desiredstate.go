package model

import (
	"fmt"
	"strings"
)

// DesiredState is the data structure that describes the desired state to apply to the FDW database
type DesiredState struct {
	// Extensions is a list of Postgres extensions (e.g. postgres_fdw)
	Extensions []Extension `yaml:"Extensions,omitempty" json:"Extensions,omitempty"`
	// Servers is a list of foreign servers
	Servers []ForeignServer `yaml:"Servers,omitempty" json:"Servers,omitempty"`
}

func (d DesiredState) String() string {
	extensionNames := make([]string, len(d.Extensions))
	for idx, extension := range d.Extensions {
		extensionNames[idx] = extension.Name
	}
	serverNames := make([]string, len(d.Servers))
	for idx, server := range d.Servers {
		serverNames[idx] = server.Name
	}
	return fmt.Sprintf(
		"extensions: {%s}, servers: {%s}",
		strings.Join(extensionNames, ","),
		strings.Join(serverNames, ","),
	)
}
