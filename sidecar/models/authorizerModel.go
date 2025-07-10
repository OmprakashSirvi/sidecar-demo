package models

import "github.com/casbin/casbin/v2"

// Authorizer for service
type BasicAuthorizer struct {
	Enforcer *casbin.SyncedEnforcer
}
