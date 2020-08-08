package model

// DesiredState is the data structure that describes the desired state to apply to the FDW database
type DesiredState struct {
	// Extensions is a list of Postgres extensions (e.g. postgres_fdw)
	Extensions []Extension `yaml:"Extensions,omitempty" json:"Extensions,omitempty"`
	// Servers is a list of foreign servers
	Servers []ForeignServer `yaml:"Servers,omitempty" json:"Servers,omitempty"`
}
