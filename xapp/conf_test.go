package xapp

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/daodao97/xgo/xredis"
)

func TestInitConf(t *testing.T) {
	var conf struct {
		Bind  string           `yaml:"bind" env:"BIND"`
		Redis []xredis.Options `yaml:"redis" envPrefix:"REDIS"`
	}
	os.Setenv("BIND", "127.0.0.1:8080")
	os.Setenv("REDIS_0_NAME", "test")
	os.Setenv("REDIS_0_ADDR", "127.0.0.1:6379")
	os.Setenv("REDIS_0_PASSWORD", "123456")
	os.Setenv("REDIS_0_DB", "0")

	// InitConf(&conf)

	env.Parse(&conf)

	fmt.Println(conf)
	for {
		time.Sleep(time.Second)
		fmt.Println(conf)
	}
}
