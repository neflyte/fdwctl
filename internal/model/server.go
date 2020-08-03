package model

type ForeignServer struct {
	Name    string
	Host    string
	Port    int
	DB      string
	Wrapper string
	Owner   string
}

func (fs *ForeignServer) Equals(fserver ForeignServer) bool {
	return fs.Name == fserver.Name && fs.Host == fserver.Host && fs.Port == fserver.Port &&
		fs.DB == fserver.DB && fs.Wrapper == fserver.Wrapper
}
