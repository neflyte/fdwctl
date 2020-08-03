package model

type UserMap struct {
	ServerName     string
	LocalUser      string
	RemoteUser     string
	RemotePassword string
}

func (um *UserMap) Equals(umap UserMap) bool {
	return um.ServerName == umap.ServerName && um.LocalUser == umap.LocalUser &&
		um.RemoteUser == umap.RemoteUser && um.RemotePassword == umap.RemotePassword
}
