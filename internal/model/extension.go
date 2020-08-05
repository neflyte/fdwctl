package model

type Extension struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"-" json:"-"`
}

func (ex *Extension) Equals(ext Extension) bool {
	return ex.Name == ext.Name
}
