package xrequest

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

var envNoProxyUpper = os.Getenv("NO_PROXY")
var envNoProxyLower = os.Getenv("no_proxy")

var (
	defaultClient     *http.Client
	defaultClientOnce sync.Once
	proxyClients      = make(map[string]*http.Client)
	proxyClientsMu    sync.RWMutex
)

// http proxy , support http, https, socks5
func GetProxyClient(proxy string) *http.Client {
	proxyUrl, err := url.Parse(proxy)
	if err != nil {
		return nil
	}
	proxyClientsMu.RLock()
	client, ok := proxyClients[proxy]
	proxyClientsMu.RUnlock()
	if ok {
		return client
	}
	proxyClientsMu.Lock()
	defer proxyClientsMu.Unlock()
	if client, ok = proxyClients[proxy]; ok {
		return client
	}
	client = &http.Client{
		Transport: &http.Transport{Proxy: proxyWithNoProxy(proxyUrl)},
	}
	proxyClients[proxy] = client
	return client
}

func GetDefaultProxyClient() *http.Client {
	// 使用单例 client，避免每次请求创建新 Transport 导致 "too many open files"
	defaultClientOnce.Do(func() {
		defaultClient = &http.Client{
			Transport: &http.Transport{Proxy: http.ProxyFromEnvironment},
		}
	})
	return defaultClient
}

// 构造一个支持 NO_PROXY/no_proxy 的 Proxy 函数
func proxyWithNoProxy(proxyURL *url.URL) func(*http.Request) (*url.URL, error) {
	noProxy := buildNoProxyList()
	return func(req *http.Request) (*url.URL, error) {
		host := req.URL.Hostname()
		if shouldBypassProxy(host, noProxy) {
			return nil, nil
		}
		return proxyURL, nil
	}
}

func buildNoProxyList() []string {
	// 合并大小写两种环境变量，逗号分隔
	merged := envNoProxyUpper
	if merged == "" {
		merged = envNoProxyLower
	} else if envNoProxyLower != "" {
		merged = merged + "," + envNoProxyLower
	}

	if merged == "" {
		return nil
	}

	parts := strings.Split(merged, ",")
	var list []string
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			list = append(list, v)
		}
	}
	return list
}

// 简单的 NO_PROXY 匹配：
// - '*' 全部直连
// - 精确匹配主机名
// - 后缀匹配域名（如 '.example.com' 或 'example.com' 命中 foo.example.com）
func shouldBypassProxy(host string, noProxyList []string) bool {
	if host == "" || len(noProxyList) == 0 {
		return false
	}
	for _, rule := range noProxyList {
		r := strings.TrimSpace(rule)
		if r == "" {
			continue
		}
		if r == "*" {
			return true
		}
		// 去掉可能的端口
		if i := strings.IndexByte(r, ':'); i >= 0 {
			r = r[:i]
		}
		// 归一化前导点
		r = strings.TrimPrefix(r, ".")
		if strings.EqualFold(host, r) {
			return true
		}
		if len(host) > len(r) && strings.HasSuffix(strings.ToLower(host), "."+strings.ToLower(r)) {
			return true
		}
	}
	return false
}
