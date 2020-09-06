package model

// ForeignServer represents a Postgres foreign server including related user mappings and remote schemas
type ForeignServer struct {
	// Name is the display name of the foreign server
	Name string `yaml:"name" json:"name" db:"foreign_server_name"`
	// Host is the hostname of the foreign server
	Host string `yaml:"host" json:"host" db:"hostname"`
	// Port is the port number of the foreign server
	Port int `yaml:"port" json:"port" db:"port"`
	// DB is the name of the database on the foreign server
	DB string `yaml:"db" json:"db" db:"dbname"`
	// Wrapper is the name of the foreign data wrapper to use with this foreign server
	Wrapper string `yaml:"wrapper,omitempty" json:"wrapper,omitempty" db:"foreign_data_wrapper_name"`
	// Owner is the owner of the foreign server
	Owner string `yaml:"-" json:"-" db:"authorization_identifier"`
	// UserMaps is the list of user mappings to create using this foreign server
	UserMaps []UserMap `yaml:"UserMap,omitempty" json:"UserMap,omitempty" db:"-"`
	// Schemas is the list of remote schemas to import from this foreign server
	Schemas []Schema `yaml:"Schemas,omitempty" json:"Schemas,omitempty" db:"-"`
}

// Equals determines if this object is equal to the supplied object
func (fs *ForeignServer) Equals(fserver ForeignServer) bool {
	return fs.Name == fserver.Name && fs.Host == fserver.Host && fs.Port == fserver.Port &&
		fs.DB == fserver.DB && fs.Wrapper == fserver.Wrapper
}
