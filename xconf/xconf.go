package xconf

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"github.com/daodao97/xgo/xlog"
)

func Init(configFile string, dest any) error {
	xlog.Debug("config file", xlog.String("path", configFile))
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	err = viper.Unmarshal(&dest)
	if err != nil {
		return err
	}

	viper.OnConfigChange(func(in fsnotify.Event) {
		err := viper.Unmarshal(&dest)
		if err != nil {
			xlog.Error("reload config error", err)
		}
		xlog.Info("reload config")
	})

	viper.WatchConfig()

	return nil
}
