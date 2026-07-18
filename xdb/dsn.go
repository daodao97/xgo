package xdb

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

func normalizeDatabaseDSN(driver, dsn string) (string, error) {
	if !strings.EqualFold(driver, "mysql") || !strings.HasPrefix(strings.ToLower(dsn), "mysql://") {
		return dsn, nil
	}

	return mysqlURLToDSN(dsn)
}

func mysqlURLToDSN(dsnURL string) (string, error) {
	u, err := url.Parse(dsnURL)
	if err != nil {
		// Do not include the URL or the parser error because either may contain
		// credentials from the original DSN.
		return "", fmt.Errorf("parse mysql URL DSN: invalid URL")
	}
	if !strings.EqualFold(u.Scheme, "mysql") {
		return "", fmt.Errorf("parse mysql URL DSN: unsupported scheme %q", u.Scheme)
	}
	if u.Hostname() == "" {
		return "", fmt.Errorf("parse mysql URL DSN: host is required")
	}
	if u.Fragment != "" {
		return "", fmt.Errorf("parse mysql URL DSN: fragment is not supported")
	}

	port := u.Port()
	if port == "" {
		port = "3306"
	}

	opt := mysqlDriver.NewConfig()
	opt.Net = "tcp"
	opt.Addr = net.JoinHostPort(u.Hostname(), port)
	opt.DBName = strings.TrimPrefix(u.Path, "/")
	if u.User != nil {
		opt.User = u.User.Username()
		opt.Passwd, _ = u.User.Password()
	}

	query := u.Query()
	if len(query) > 0 {
		opt.Params = make(map[string]string, len(query))
	}
	for key, values := range query {
		if len(values) > 0 {
			opt.Params[key] = values[len(values)-1]
		}
	}

	return opt.FormatDSN(), nil
}
