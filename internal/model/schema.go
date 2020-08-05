package model

type Schema struct {
	ServerName     string `yaml:"-" json:"-"`
	LocalSchema    string `yaml:"localschema" json:"localschema"`
	RemoteSchema   string `yaml:"remoteschema" json:"remoteschema"`
	ImportENUMs    bool   `yaml:"importenums" json:"importenums"`
	ENUMConnection string `yaml:"enumconnection,omitempty" json:"enumconnection,omitempty"`
}
