package xapp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env"
	"github.com/daodao97/xgo/xlog"
	"github.com/daodao97/xgo/xutil"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

// conf.yaml + conf.{env}.yaml

var confDir = []string{
	".",
	"../",
}

func SetConfDir(dir ...string) {
	confDir = append(confDir, dir...)
}

func getConfFile() []string {
	files := []string{
		"conf.yaml",
	}
	env := Args.AppEnv
	if env != "" {
		files = append(files, fmt.Sprintf("conf.%s.yaml", strings.ToLower(env)))
	}
	return files
}

func fileConf(dest any, files ...string) error {
	for _, dir := range confDir {
		for _, file := range files {
			// 检查文件是否存在
			if _, err := os.Stat(filepath.Join(dir, file)); os.IsNotExist(err) {
				continue
			}

			// 读取文件内容
			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("读取配置文件失败 %s: %v", file, err)
			}

			// 解析 YAML 到目标结构体
			if err := yaml.Unmarshal(data, dest); err != nil {
				return fmt.Errorf("解析配置文件失败 %s: %v", file, err)
			}

			// 处理环境变量替换
			if err := env.Parse(dest); err != nil {
				return fmt.Errorf("处理环境变量失败: %v", err)
			}

			xlog.Info("load config", xlog.String("file", file))
		}
	}

	return nil
}

// InitConf 初始化配置
func InitConf(dest any) error {
	if !xutil.IsPtr(dest) {
		return fmt.Errorf("配置目标必须是结构体指针类型")
	}
	confFiles := getConfFile()
	err := fileConf(dest, confFiles...)
	if err != nil {
		return err
	}
	xlog.Info("init config", xlog.String("files", strings.Join(confFiles, ",")))
	// 创建文件监听器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("创建监听器失败: %v", err)
	}

	// 启动监听协程
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					xlog.Error("监听器退出")
					return
				}
				// 如果是写入或创建事件，重新加载配置
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					if err := fileConf(dest); err != nil {
						xlog.Error("重新加载配置失败", xlog.Err(err))
					} else {
						xlog.Info("配置已重新加载", xlog.String("file", event.Name))
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				xlog.Error("监听错误", xlog.Err(err))
			}
		}
	}()

	for _, file := range confFiles {
		if err := watcher.Add(file); err != nil {
			// 如果文件不存在，跳过
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("添加文件监听失败 %s: %v", file, err)
		}
	}

	return nil
}
