package utils

import (
	"bytes"
	"encoding/json"
	"net"
	"net/url"
	"os"
	"strings"
)

func JsonStr(v any) string {
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	jsonEncoder.Encode(v)
	return bf.String()
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
