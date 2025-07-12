package handlers

import (
	"net/http/httputil"
	"sidecar/models"

	"github.com/gin-gonic/gin"
)

// This will be the handler for all the requests
func ProxyRequestHandler(proxy *httputil.ReverseProxy, route models.ProxyRoute) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

