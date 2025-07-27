package xrequest

import (
	"net/http"
	"net/url"
	"os"
)

var envHttpProxy = os.Getenv("HTTP_PROXY")
var envHttpsProxy = os.Getenv("HTTPS_PROXY")
var envAllProxy = os.Getenv("ALL_PROXY")

// http proxy , support http, https, socks5
func GetProxyClient(proxy string) *http.Client {
	proxyUrl, err := url.Parse(proxy)
	if err != nil {
		return nil
	}
	return &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
	}
}

func GetDefaultProxyClient() *http.Client {
	proxy := getEnvProxy()
	if proxy == "" {
		return &http.Client{
			Transport: &http.Transport{},
		}
	}
	return GetProxyClient(proxy)
}

func getEnvProxy() string {
	proxy := envHttpProxy
	if proxy == "" {
		proxy = envHttpsProxy
	}
	if proxy == "" {
		proxy = envAllProxy
	}
	return proxy
}
