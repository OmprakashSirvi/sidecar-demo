package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
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
	// Can modify the response, and see to the response and give out an error if something does not seems right
	// Can be used in logging and tracking
	return nil
}
