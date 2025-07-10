package models

// Authz config for loading RBAC configuration files
type AuthzConfig struct {
	// Type of RBAC
	AuthzType string `mapstructure:"type"`
	// Name/location of the model file
	ModelFile string `mapstructure:"model-file"`
	// Name/location of policy file
	PolicyFile string `mapstructure:"policy-file"`
}
