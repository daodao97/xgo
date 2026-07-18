package xdb

import (
	"testing"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

func TestNormalizeDatabaseDSNConvertsMySQLURL(t *testing.T) {
	input := "mysql://appuser:1d834b9e4117b151a2d7491e@paas-addon-mysql-c7ntw7219m:3306/appdb"

	dsn, err := normalizeDatabaseDSN("mysql", input)
	if err != nil {
		t.Fatalf("normalizeDatabaseDSN failed: %v", err)
	}

	conf, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("converted DSN is invalid: %v", err)
	}
	if conf.User != "appuser" || conf.Passwd != "1d834b9e4117b151a2d7491e" {
		t.Fatalf("unexpected credentials: user %q password %q", conf.User, conf.Passwd)
	}
	if conf.Net != "tcp" || conf.Addr != "paas-addon-mysql-c7ntw7219m:3306" {
		t.Fatalf("unexpected network address: %s(%s)", conf.Net, conf.Addr)
	}
	if conf.DBName != "appdb" {
		t.Fatalf("unexpected database name %q", conf.DBName)
	}
}

func TestNormalizeDatabaseDSNHandlesDefaultsEscapingAndParams(t *testing.T) {
	input := "mysql://app%40user:p%40ss@mysql.internal/app%20db?parseTime=true&charset=utf8mb4"

	dsn, err := normalizeDatabaseDSN("mysql", input)
	if err != nil {
		t.Fatalf("normalizeDatabaseDSN failed: %v", err)
	}

	conf, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("converted DSN is invalid: %v", err)
	}
	if conf.User != "app@user" || conf.Passwd != "p@ss" {
		t.Fatalf("unexpected decoded credentials: user %q password %q", conf.User, conf.Passwd)
	}
	if conf.Addr != "mysql.internal:3306" {
		t.Fatalf("expected default port, got %q", conf.Addr)
	}
	if conf.DBName != "app db" {
		t.Fatalf("unexpected decoded database name %q", conf.DBName)
	}
	if !conf.ParseTime || conf.Params["charset"] != "utf8mb4" {
		t.Fatalf("unexpected query parameters: parseTime=%v params=%v", conf.ParseTime, conf.Params)
	}
}

func TestNormalizeDatabaseDSNLeavesNativeAndOtherDriversUnchanged(t *testing.T) {
	tests := []struct {
		name   string
		driver string
		dsn    string
	}{
		{name: "native mysql", driver: "mysql", dsn: "user:pass@tcp(mysql:3306)/app"},
		{name: "postgres URL", driver: "postgres", dsn: "postgres://user:pass@db/app"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normalizeDatabaseDSN(test.driver, test.dsn)
			if err != nil {
				t.Fatalf("normalizeDatabaseDSN failed: %v", err)
			}
			if got != test.dsn {
				t.Fatalf("expected DSN to remain unchanged, got %q", got)
			}
		})
	}
}

func TestNormalizeDatabaseDSNRejectsMissingHost(t *testing.T) {
	if _, err := normalizeDatabaseDSN("mysql", "mysql:///appdb"); err == nil {
		t.Fatal("expected missing host error")
	}
}
