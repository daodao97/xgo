package xproxy

import (
	"strings"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/gin-gonic/gin"
)

var filteredHeaders = map[string]bool{
	"x-real-ip":               true,
	"x-forwarded-for":         true,
	"x-forwarded-proto":       true,
	"x-forwarded-port":        true,
	"x-forwarded-host":        true,
	"x-forwarded-server":      true,
	"cf-warp-tag-id":          true,
	"cf-visitor":              true,
	"cf-ray":                  true,
	"cf-connecting-ip":        true,
	"cf-ipcountry":            true,
	"cdn-loop":                true,
	"remote-host":             true,
	"x-frame-options":         true,
	"x-xss-protection":        true,
	"content-security-policy": true,
	"content-encoding":        true,
	"origin":                  true,
	"referer":                 true,
}

func Request(client tls_client.HttpClient, endpoint string, c *gin.Context) (*http.Response, error) {
	targetURL := endpoint + c.Request.URL.Path
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	// Create request
	req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		return nil, err
	}

	// Add origin header and referer header
	req.Header.Add("Origin", endpoint)
	req.Header.Add("Referer", endpoint)

	// Add CORS headers
	req.Header.Add("Access-Control-Allow-Origin", "*")
	req.Header.Add("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	req.Header.Add("Access-Control-Allow-Headers", "Content-Type, Authorization")
	req.Header.Add("Access-Control-Allow-Credentials", "true")

	for name, values := range c.Request.Header {
		if _, found := filteredHeaders[strings.ToLower(name)]; found {
			continue
		}
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Copy headers
	for name, values := range resp.Header {
		if _, found := filteredHeaders[strings.ToLower(name)]; found {
			continue
		}
		for _, value := range values {
			if strings.ToLower(name) == "content-length" {
				continue
			}
			if strings.ToLower(name) == "set-cookie" {
				cookies := strings.Split(value, ";")
				newCookies := make([]string, 0)
				for _, cookie := range cookies {
					cookieParts := strings.SplitN(cookie, "=", 2)
					if len(cookieParts) == 2 && strings.ToLower(strings.TrimSpace(cookieParts[0])) != "domain" {
						newCookies = append(newCookies, cookie)
					}
				}
				c.Writer.Header().Add("Set-Cookie", strings.Join(newCookies, ";"))
			} else {
				c.Writer.Header().Add(name, value)
			}
		}
	}

	// Remove CSP
	resp.Header.Del("Content-Security-Policy")
	return resp, nil
}
