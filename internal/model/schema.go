package model

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
}
