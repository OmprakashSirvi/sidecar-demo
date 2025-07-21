package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"sidecar/applogger"
	"sidecar/globals"
	"sidecar/models"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// This middleware will only be added to a route when route policies are given..
func AuthorizeRequest(route *models.ProxyRoute) gin.HandlerFunc {
	// Now check if there are any resource match conditions in the route policies..
	// If they are enabled then.. check if super override is true
	// Use suitable middlewares based on these conditions..
	resourceMatchEnabled := false
	for _, policy := range route.RoutePolicies {
		// TODO: Probably should create a helper function here..
		if policy.ResourceMatches != (models.ResourceMatch{}) {
			resourceMatchEnabled = true
			break
		}
	}

	return func(ctx *gin.Context) {
		ctxLogger := applogger.GetCtxLogger(ctx)
		userId := getUserIdFromHeader(ctx)
		path := ctx.Request.URL.Path
		method := ctx.Request.Method
		logger := ctxLogger.With().Str("userId", userId).Str("method", method).Str("path", path).Logger()

		if !resourceMatchEnabled || route.SuperOverride {
			// Just authorize this request on basis of route access.
			status, err := checkUserAccessPermissions(userId, path, method, &logger)
			if err != nil && !route.SuperOverride {
				ctx.AbortWithStatusJSON(status, err)
				return
			}
			if err == nil {
				ctx.Next()
			}
		}

		// If we are here, then the user does not have super permission to the route, and resource match is enabled
		// Check resource permissions
		for _, policy := range route.RoutePolicies {
			resourceMatch := policy.ResourceMatches
			switch resourceMatch.From {
			case "param":
				// Extract the resourceId (resourceMatch.Name) from param
				resourceId := ctx.Param(resourceMatch.Name)
				if resourceId == "" {
					// No resource Id found for this resource name..
					errMsg := fmt.Sprintf("no resource id provided for policy name: %v", policy.Name)
					logger.Error().Msg(errMsg)
					ctx.AbortWithStatusJSON(http.StatusForbidden, errMsg)
					return
				}

				if resourceId != userId {
					// Here in this case the current resource-token does not match the resource-id
					errMsg := fmt.Sprintf("the resource id does match with the resourceId %v", resourceId)
					logger.Error().Msg(errMsg)
					ctx.AbortWithStatusJSON(http.StatusForbidden, errMsg)
					return
				}
			case "query":
				// Extract the resourceId (resourceMatch.Name) from query
				// TODO: Will implement this later..
			default:
				// This configuration is not supported, probably should return from here..
			}
		}

	}
}

// This checks if the user provided in the context header has access to the given endpoint or not
func checkUserAccessPermissions(userId string, path string, method string, parentLogger *zerolog.Logger) (int, error) {
	logger := parentLogger.With().Str("function", "checkUserAccessPermissions").Logger()

	// This checks if the user has relevant permissions to access this endpoint
	ok, err := globals.Global.UserAuthorizer.Enforcer.Enforce(userId, path, method)
	if err != nil {
		msg := "something went wrong while authorizing user"
		logger.Error().Err(err).Msg(msg)
		return http.StatusBadRequest, err
	}
	if !ok {
		errMsg := errors.New("user not allowed to access this route")
		logger.Error().Msg(errMsg.Error())
		return http.StatusForbidden, errMsg
	}

	logger.Debug().Msg("allowed access")
	// This is a bad design I guess, should not return 0
	return 0, nil
}

// TODO: This has been delegated to a new middleware, which will verify the jwt token
// Based on the token type it will add a new header x-{tokenType}-id
// TODO: This should be enhanced to take care of different types of tokens other than just userId
// TODO: This will be enhanced to verify the jwt, introspect the token to get userID
// TODO: We should also consider configuring where the auth-token will be provided..
func getUserIdFromHeader(c *gin.Context) string {
	return c.GetHeader("x-user-token")
}
