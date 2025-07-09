package globals

import "github.com/casbin/casbin/v2"

type globalVars struct {
	// Backend service URL
	// Need to setup reverse proxy for this and forward request to sidecar here..
	ProxyBackend string
	ConfigDir    string
	AuthzConfigs []AuthzConfig

	UserAuthorizer    *BasicAuthorizer
	ServiceAuthorizer *BasicAuthorizer
}

type BasicAuthorizer struct {
	Enforcer *casbin.SyncedEnforcer
}

type AuthzConfig struct {
	AuthzType  string `mapstructure:"type"`
	ModelFile  string `mapstructure:"model-file"`
	PolicyFile string `mapstructure:"policy-file"`
}

var Global = globalVars{}
