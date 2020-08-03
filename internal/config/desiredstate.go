package config

type DesiredStateSecretK8s struct {
	Namespace  string `yaml:"namespace" json:"namespace"`
	SecretName string `yaml:"secretName" json:"secretName"`
	SecretKey  string `yaml:"secretKey" json:"secretKey"`
}

type DesiredStateSecret struct {
	Value         string                `yaml:"value,omitempty" json:"value,omitempty"`
	FromEnv       string                `yaml:"fromEnv,omitempty" json:"fromEnv,omitempty"`
	FromFile      string                `yaml:"fromFile,omitEmpty" json:"fromFile,omitempty"`
	FromK8sSecret DesiredStateSecretK8s `yaml:"fromK8s,omitempty" json:"fromK8s,omitempty"`
}

type DesiredStateSchema struct {
	LocalSchema    string `yaml:"localschema" json:"localschema"`
	RemoteSchema   string `yaml:"remoteschema" json:"remoteschema"`
	ImportENUMs    bool   `yaml:"importenums" json:"importenums"`
	ENUMConnection string `yaml:"enumconnection,omitempty" json:"enumconnection,omitempty"`
}

type DesiredStateUserMap struct {
	LocalUser      string `yaml:"localuser" json:"localuser"`
	RemoteUser     string `yaml:"remoteuser" json:"remoteuser"`
	RemotePassword string `yaml:"remotepassword" json:"remotepassword"`
}

type DesiredStateServer struct {
	Name    string                `yaml:"name" json:"name"`
	Host    string                `yaml:"host" json:"host"`
	Port    int                   `yaml:"port" json:"port"`
	DB      string                `yaml:"db" json:"db"`
	Wrapper string                `yaml:"wrapper,omitempty" json:"wrapper,omitempty"`
	UserMap []DesiredStateUserMap `yaml:"UserMap,omitempty" json:"UserMap,omitempty"`
	Schemas []DesiredStateSchema  `yaml:"Schemas,omitempty" json:"Schemas,omitempty"`
}

type DesiredState struct {
	Servers []DesiredStateServer `yaml:"Servers,omitempty"`
}
