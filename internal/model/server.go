package model

type ForeignServer struct {
	Name     string    `yaml:"name" json:"name"`
	Host     string    `yaml:"host" json:"host"`
	Port     int       `yaml:"port" json:"port"`
	DB       string    `yaml:"db" json:"db"`
	Wrapper  string    `yaml:"wrapper,omitempty" json:"wrapper,omitempty"`
	Owner    string    `yaml:"-" json:"-"`
	UserMaps []UserMap `yaml:"UserMap,omitempty" json:"UserMap,omitempty"`
	Schemas  []Schema  `yaml:"Schemas,omitempty" json:"Schemas,omitempty"`
}

func (fs *ForeignServer) Equals(fserver ForeignServer) bool {
	return fs.Name == fserver.Name && fs.Host == fserver.Host && fs.Port == fserver.Port &&
		fs.DB == fserver.DB && fs.Wrapper == fserver.Wrapper
}
