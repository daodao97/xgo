package conf

import (
	"log/slog"
	"os"

	"github.com/daodao97/xgo/utils"
	"github.com/daodao97/xgo/xdb"
	"github.com/go-redis/redis/v8"

	"github.com/daodao97/xgo/xconf"
	"github.com/daodao97/xgo/xlog"
)

type config struct {
	ServerAddr string         `yaml:"server_addr"`
	AdminPath  string         `yaml:"admin_path"`
	Database   *xdb.Config    `yaml:"database"`
	Redis      *redis.Options `yaml:"redis"`
	JwtSecret  string         `yaml:"jwt_secret"`
}

func (c *config) Print() {
	xlog.Debug("load config", slog.Any("config", c))
}

var _c *config

func Get() *config {
	return _c
}

func InitConf() error {
	_c = &config{
		ServerAddr: ":3000",
		AdminPath:  "/_",
		JwtSecret:  "hsxtypr",
	}

	if err := xconf.Init("./conf.yaml", &_c); err != nil {
		return nil
	}

	if redisEnv := os.Getenv("REDIS_ADDR"); redisEnv != "" {
		_c.Redis.Addr = redisEnv
	}

	if !utils.IsGoRun() {
		xlog.SetLogger(xlog.StdoutJson())
	}

	_c.Print()

	return nil
}
