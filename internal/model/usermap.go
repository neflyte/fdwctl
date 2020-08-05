package model

type UserMap struct {
	ServerName   string `yaml:"-" json:"-"`
	LocalUser    string `yaml:"localuser" json:"localuser"`
	RemoteUser   string `yaml:"remoteuser" json:"remoteuser"`
	RemoteSecret Secret `yaml:"remotesecret" json:"remotesecret"`
}

func (um *UserMap) Equals(umap UserMap) bool {
	return um.LocalUser == umap.LocalUser && um.RemoteUser == umap.RemoteUser && um.RemoteSecret.Equals(umap.RemoteSecret)
}
