package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func NewReverseProxy(upstream string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(upstream)
	if err != nil {
		return nil, err
	}

	// ginProxy := &httputil.ReverseProxy {
	// 	ModifyResponse: UpstreamResponseModifier,
	// }
	proxy := httputil.NewSingleHostReverseProxy(url)
	// We can add metrics of some other modifications on the response from the backend service
	proxy.ModifyResponse = UpstreamResponseModifier

	return proxy, nil
}

func UpstreamResponseModifier(r *http.Response) error {
	fmt.Printf("got response from backend service: %v\n", r.StatusCode)
	return nil
}

// panic if something goes wrong
func setProxyRoutes(router *gin.Engine, proxy *httputil.ReverseProxy) {
	// This will be a good location to get the proxy routes
	if ok := viper.IsSet("proxy-routes"); !ok {
		fmt.Println("proxy-routes is not set, hence not configuring any routes")
		return
	}

	var routes []Route
	err := viper.UnmarshalKey("proxy-routes", &routes)
	if err != nil {
		panic("invalid proxy-routes configuration")
	}

	for _, route := range routes {
		// TODO: Here need to add more switch cases
		switch route.Type {
		case "GET":
			router.GET(route.Path, proxyRequestHandler(proxy, route))
		}
	}
}

func proxyRequestHandler(proxy *httputil.ReverseProxy, route Route) gin.HandlerFunc {
	fmt.Printf("handling route for: %v:%v\n", route.Type, route.Path)
	return func(ctx *gin.Context) {
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
