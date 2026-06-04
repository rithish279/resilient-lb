package lb

import (
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Backend struct {
	URL 	*url.URL
	Healthy	atomic.Bool
	proxy	*httputil.ReverseProxy
}

func NewBackend(rawURL string) *Backend {
	u := mustParseURL(rawURL)
	b := &Backend{
		URL: u,
		proxy: httputil.NewSingleHostReverseProxy(u),
	}
	
	return b
}

func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}