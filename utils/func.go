package utils

import (
	"encoding/json"
	"net"
	"net/url"
	"os"
	"strings"
)

func JsonStr(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func InArray[T comparable](t T, arr []T) bool {
	for _, v := range arr {
		if v == t {
			return true
		}
	}
	return false
}

func IsUrl(str string) bool {
	url, err := url.ParseRequestURI(str)
	if err != nil {
		return false
	}

	address := net.ParseIP(url.Host)
	if address == nil {
		return strings.Contains(url.Host, ".")
	}

	return true
}

func IsGoRun() (withGoRun bool) {
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		withGoRun = true
	} else {
		withGoRun = false
	}
	return
}
