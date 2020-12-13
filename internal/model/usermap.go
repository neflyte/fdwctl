package model

// UserMap represents a Postgres user mapping
type UserMap struct {
	// ServerName is the name of the foreign server
	ServerName string `yaml:"-" json:"-"`
	// LocalUser is the name of the local database user to map
	LocalUser string `yaml:"localuser" json:"localuser"`
	// RemoteUser is the name of the remote database user to connect as
	RemoteUser string `yaml:"remoteuser" json:"remoteuser"`
	// RemoteSecret configures how to retrieve the optional credential for the RemoteUser user
	RemoteSecret Secret `yaml:"remotesecret" json:"remotesecret"`
}

// Equals determines if this object is equal to the supplied object
func (um *UserMap) Equals(umap UserMap) bool {
	return um.LocalUser == umap.LocalUser && um.RemoteUser == umap.RemoteUser && um.RemoteSecret.Value == umap.RemoteSecret.Value
}
