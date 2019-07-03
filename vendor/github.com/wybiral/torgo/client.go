package torgo

import (
	"net/http"
	"net/url"

	"golang.org/x/net/proxy"
)

// NewClient return new HTTP client that uses a Tor SOCKS proxy.
// `addr` is the host:port address of SOCKS proxy (usually "127.0.0.1:9050")
func NewClient(addr string) (*http.Client, error) {
	u, err := url.Parse("socks5://" + addr)
	if err != nil {
		return nil, err
	}
	d, err := proxy.FromURL(u, proxy.Direct)
	if err != nil {
		return nil, err
	}
	t := &http.Transport{Dial: d.Dial}
	return &http.Client{Transport: t}, nil
}
