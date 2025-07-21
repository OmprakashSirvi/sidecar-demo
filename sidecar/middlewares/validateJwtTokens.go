package middlewares

import (
	"sidecar/models"

	"github.com/gin-gonic/gin"
)

func ValidateJwtTokens(route *models.ProxyRoute) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, policy := range route.RoutePolicies {
			if policy.Type != "jwt" {
				continue
			}

			// TODO: Here need to validate the tokens only if they are a valid tokens
		}
	}
}
