package model

// Extension represents a Postgres extension
type Extension struct {
	Name    string `yaml:"name" json:"name" db:"extname"`
	Version string `yaml:"-" json:"-" db:"extversion"`
}

// Equals determines if this object is equal to the supplied object
func (ex *Extension) Equals(ext Extension) bool {
	return ex.Name == ext.Name
}
