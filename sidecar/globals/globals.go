package globals

import "github.com/casbin/casbin/v2"

type globalVars struct {
	// Backend service URL
	// Need to setup reverse proxy for this and forward request to sidecar here..
	ProxyBackend string

	UserAuthorizer *BasicAuthorizer
}

type BasicAuthorizer struct {
	Enforcer *casbin.SyncedEnforcer
}

var Global = globalVars{}
