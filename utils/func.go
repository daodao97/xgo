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

	host := url.Hostname()
	if host == "" {
		return false
	}

	address := net.ParseIP(host)
	if address == nil {
		if !strings.Contains(host, ".") {
			return false
		}

		parts := strings.Split(host, ".")
		allNumeric := true
		for _, part := range parts {
			if part == "" {
				return false
			}
			for _, r := range part {
				if r < '0' || r > '9' {
					allNumeric = false
					break
				}
			}
		}
		// e.g. 192.168.1 is not a valid host/ip
		if allNumeric {
			return false
		}
		return true
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
