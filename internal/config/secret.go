package config

type SecretK8s struct {
	Namespace  string `yaml:"namespace" json:"namespace"`
	SecretName string `yaml:"secretName" json:"secretName"`
	SecretKey  string `yaml:"secretKey" json:"secretKey"`
}

type Secret struct {
	Value         string    `yaml:"value,omitempty" json:"value,omitempty"`
	FromEnv       string    `yaml:"fromEnv,omitempty" json:"fromEnv,omitempty"`
	FromFile      string    `yaml:"fromFile,omitEmpty" json:"fromFile,omitempty"`
	FromK8sSecret SecretK8s `yaml:"fromK8s,omitempty" json:"fromK8s,omitempty"`
}
