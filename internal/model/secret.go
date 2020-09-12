package model

// SecretK8s represents the location of a base64-encoded credential in a Kubernetes secret
type SecretK8s struct {
	// Namespace is the Kubernetes namespace that contains the secret
	Namespace string `yaml:"namespace" json:"namespace"`
	// SecretName is the name of the Kubernetes secret object
	SecretName string `yaml:"secretName" json:"secretName"`
	// SecretKey is the name of the key underneath .data which contains the credential
	SecretKey string `yaml:"secretKey" json:"secretKey"`
}

// Equals determines if this object is equal to the supplied object
func (sk *SecretK8s) Equals(secret SecretK8s) bool {
	return secret.Namespace == sk.Namespace && secret.SecretName == sk.SecretName && secret.SecretKey == sk.SecretKey
}

// Secret defines where to retrieve a credential from
type Secret struct {
	// Value represents an explicit credential value to be used verbatim
	Value string `yaml:"value,omitempty" json:"value,omitempty" db:"password,omitempty"`
	// FromEnv represents an environment variable to read the credential from
	FromEnv string `yaml:"fromEnv,omitempty" json:"fromEnv,omitempty" db:"-"`
	// FromFile represents a path and filename to read the credential from
	FromFile string `yaml:"fromFile,omitempty" json:"fromFile,omitempty" db:"-"`
	// FromK8sSecret represents a Kubernetes secret to read the credential from
	FromK8sSecret SecretK8s `yaml:"fromK8s,omitempty" json:"fromK8s,omitempty" db:"-"`
}

// Equals determines if this object is equal to the supplied object
func (s *Secret) Equals(secret Secret) bool {
	return secret.Value == s.Value && secret.FromEnv == s.FromEnv && secret.FromFile == s.FromFile && secret.FromK8sSecret.Equals(s.FromK8sSecret)
}
