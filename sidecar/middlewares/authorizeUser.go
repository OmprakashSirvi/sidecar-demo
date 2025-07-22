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
// This middleware will only be called after the jwt-tokens have been validated/verified..
func AuthorizeRequest(route *models.ProxyRoute) gin.HandlerFunc {
	// Now check if there are any resource match conditions in the route policies..
	// If they are enabled then.. check if super override is true
	// Use suitable middlewares based on these conditions..
	resourceMatchEnabled := false
	for _, policy := range route.RoutePolicies {
		// TODO: Probably should create a helper function here..
		if policy.ResourceMatch != (models.ResourceMatch{}) {
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

		// If we are here, then the user does not have super permission to the route, and resource match is enabled
		// Check resource permissions
		if resourceMatchEnabled {
			for _, policy := range route.RoutePolicies {
				resourceMatch := policy.ResourceMatch
				resourceTokenId := getResourceIdFromHeader(ctx, resourceMatch.Name)
				switch resourceMatch.From {
				case "param":
					logger.Trace().Msg("checking the request parameter for resourceId")
					// Extract the resourceId (resourceMatch.Name) from param
					resourceId := ctx.Param(resourceMatch.Name)
					err := checkResourceMatch(resourceId, resourceTokenId, &policy, &logger)
					if err != nil && !route.SuperOverride {
						ctx.AbortWithError(http.StatusForbidden, err)
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

		// If super-override is enabled then a user with specified privileges can access this route.
		// If the user do not have access then in this case we check for resource matches..
		// Just authorize this request on basis of route access.
		// This checks the authz models and sees if the user has permissions or not..
		status, err := checkUserAccessPermissions(userId, path, method, &logger)
		if err != nil {
			ctx.AbortWithError(status, err)
			return
		}

		ctx.Next()
	}
}

func checkResourceMatch[S string](resourceId S, resourceTokenId S, policy *models.RoutePolicy, logger *zerolog.Logger) error {
	resourceMatch := policy.ResourceMatch
	if resourceId == "" {
		// No resource Id found for this resource name..
		errMsg := fmt.Errorf("no %v provided in request %v for policy name: %v", resourceMatch.Name, resourceMatch.From, policy.Name)
		logger.Error().Msg(errMsg.Error())
		return errMsg
	}

	if resourceId != resourceTokenId {
		// Here in this case the current resource-token does not match the resource-id
		errMsg := fmt.Errorf("the %v does match with the resourceId %v from request %v", resourceMatch.Name, resourceId, resourceMatch.From)
		logger.Error().Msg(errMsg.Error())
		return errMsg
	}

	return nil
}

// This checks if the user provided in the context header has access to the given endpoint or not
func checkUserAccessPermissions[S string](userId S, path S, method S, parentLogger *zerolog.Logger) (int, error) {
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
	return c.GetHeader("x-user-id")
}

func getResourceIdFromHeader(c *gin.Context, name string) string {
	return c.GetHeader(fmt.Sprintf("x-%v", name))
}
