package model

type SecretK8s struct {
	Namespace  string `yaml:"namespace" json:"namespace"`
	SecretName string `yaml:"secretName" json:"secretName"`
	SecretKey  string `yaml:"secretKey" json:"secretKey"`
}

func (sk *SecretK8s) Equals(secret SecretK8s) bool {
	return secret.Namespace == sk.Namespace && secret.SecretName == sk.SecretName && secret.SecretKey == sk.SecretKey
}

type Secret struct {
	Value         string    `yaml:"value,omitempty" json:"value,omitempty"`
	FromEnv       string    `yaml:"fromEnv,omitempty" json:"fromEnv,omitempty"`
	FromFile      string    `yaml:"fromFile,omitempty" json:"fromFile,omitempty"`
	FromK8sSecret SecretK8s `yaml:"fromK8s,omitempty" json:"fromK8s,omitempty"`
}

func (s *Secret) Equals(secret Secret) bool {
	return secret.Value == s.Value && secret.FromEnv == s.FromEnv && secret.FromFile == s.FromFile && secret.FromK8sSecret.Equals(s.FromK8sSecret)
}
