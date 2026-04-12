package xapp

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/daodao97/xgo/xenv"
)

func TestInitConfReloadsFilesFromResolvedPaths(t *testing.T) {
	tempDir := t.TempDir()
	writeFile(t, filepath.Join(tempDir, "conf.yaml"), "bind: base\n")
	writeFile(t, filepath.Join(tempDir, "conf.dev.yaml"), "bind: dev\n")

	originalConfDir := confDir
	originalAppEnv := Args.AppEnv
	originalXEnv := xenv.AppEnv
	confDir = []string{tempDir}
	Args.AppEnv = "dev"
	xenv.AppEnv = "dev"
	defer func() {
		confDir = originalConfDir
		Args.AppEnv = originalAppEnv
		xenv.AppEnv = originalXEnv
		CloseConfWatcher()
	}()

	var conf struct {
		Bind string `yaml:"bind"`
	}

	if err := InitConf(&conf); err != nil {
		t.Fatalf("InitConf failed: %v", err)
	}
	if conf.Bind != "dev" {
		t.Fatalf("expected initial dev config, got %q", conf.Bind)
	}

	writeFile(t, filepath.Join(tempDir, "conf.dev.yaml"), "bind: reloaded\n")
	waitFor(t, time.Second, func() bool {
		return conf.Bind == "reloaded"
	})
}

func TestInitConfClosesPreviousWatcherOnReinit(t *testing.T) {
	tempDir := t.TempDir()
	writeFile(t, filepath.Join(tempDir, "conf.yaml"), "bind: first\n")

	originalConfDir := confDir
	originalAppEnv := Args.AppEnv
	originalXEnv := xenv.AppEnv
	confDir = []string{tempDir}
	Args.AppEnv = "dev"
	xenv.AppEnv = "dev"
	defer func() {
		confDir = originalConfDir
		Args.AppEnv = originalAppEnv
		xenv.AppEnv = originalXEnv
		CloseConfWatcher()
	}()

	var first struct {
		Bind string `yaml:"bind"`
	}
	if err := InitConf(&first); err != nil {
		t.Fatalf("first InitConf failed: %v", err)
	}

	var second struct {
		Bind string `yaml:"bind"`
	}
	if err := InitConf(&second); err != nil {
		t.Fatalf("second InitConf failed: %v", err)
	}

	writeFile(t, filepath.Join(tempDir, "conf.yaml"), "bind: second\n")
	waitFor(t, time.Second, func() bool {
		return second.Bind == "second"
	})
	if first.Bind == "second" {
		t.Fatal("expected previous watcher to stop reloading old config target")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func waitFor(t *testing.T, timeout time.Duration, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}
