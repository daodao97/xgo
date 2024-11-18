package xapp

import (
	"fmt"
	"testing"
	"time"
)

func TestInitConf(t *testing.T) {
	var conf struct {
		Bind string `yaml:"bind" env:"BIND"`
	}

	InitConf(&conf)

	fmt.Println(conf)
	for {
		time.Sleep(time.Second)
		fmt.Println(conf)
	}
}
