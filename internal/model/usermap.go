package model

type UserMap struct {
	ServerName     string `yaml:"-" json:"-"`
	LocalUser      string `yaml:"localuser" json:"localuser"`
	RemoteUser     string `yaml:"remoteuser" json:"remoteuser"`
	RemotePassword string `yaml:"remotepassword" json:"remotepassword"`
}

func (um *UserMap) Equals(umap UserMap) bool {
	return um.LocalUser == umap.LocalUser && um.RemoteUser == umap.RemoteUser && um.RemotePassword == umap.RemotePassword
}
