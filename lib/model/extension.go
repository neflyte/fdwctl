package model

import "fmt"

// Extension represents a Postgres extension
type Extension struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"-" json:"-"`
}

// Equals determines if this object is equal to the supplied object
func (ex *Extension) Equals(ext Extension) bool {
	return ex.Name == ext.Name
}

func (ex *Extension) String() string {
	return fmt.Sprintf("name: %s, version: %s", ex.Name, ex.Version)
}
