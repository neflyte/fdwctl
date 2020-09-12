package model

// Schema represents a foreign schema configuration
type Schema struct {
	ServerName     string  `yaml:"-" json:"-" db:"foreign_server_name"`                             // ServerName is the name of the foreign server that will be used to import the remote schema
	LocalSchema    string  `yaml:"localschema" json:"localschema" db:"foreign_table_schema"`        // LocalSchema is the name of the local schema that the remote schema will be imported into
	RemoteSchema   string  `yaml:"remoteschema" json:"remoteschema" db:"remote_schema"`             // RemoteSchema is the name of the remote schema to import
	ImportENUMs    bool    `yaml:"importenums" json:"importenums" db:"-"`                           // ImportENUMs indicates whether remote ENUM types should be auto-created locally before importing
	ENUMConnection string  `yaml:"enumconnection,omitempty" json:"enumconnection,omitempty" db:"-"` // ENUMConnection specifies the connection string to the remote database for reading ENUM definitions
	ENUMSecret     Secret  `yaml:"enumsecret,omitempty" json:"enumsecret,omitempty" db:"-"`         // ENUMSecret configures how to retrieve the optional credential for the ENUMConnection connection string
	Grants         []Grant `yaml:"Grants,omitempty" json:"Grants,omitempty" db:"-"`                 // Grants is a list of permissions to grant users for the local schema
}

// Grant represents permissions to grant a user in the local schema
type Grant struct {
	User        string     `yaml:"user" json:"user"`               // User is the name of the local user to which permissions will be granted
	Permissions Permission `yaml:"Permissions" json:"Permissions"` // Permissions is a list of permissions to grant
}

// Permission represents a permission to grant a user
type Permission struct {
	Usage  bool `yaml:"usage" json:"usage"`   // Usage grants usage on a schema
	Select bool `yaml:"select" json:"select"` // Select grants SELECT permission on all tables in a schema
}
