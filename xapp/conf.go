package xapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/daodao97/xgo/xenv"
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

type confWatcherState struct {
	cancel context.CancelFunc
	done   chan struct{}
}

var (
	confWatcherMu    sync.Mutex
	currentConfWatch *confWatcherState
)

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

func fileConf(dest any, files ...string) ([]string, error) {
	var loadedFiles []string
	for _, dir := range confDir {
		for _, file := range files {
			// 检查文件是否存在
			if _, err := os.Stat(filepath.Join(dir, file)); os.IsNotExist(err) {
				continue
			}

			absFile, err := filepath.Abs(filepath.Join(dir, file))
			if err != nil {
				return nil, fmt.Errorf("获取绝对路径失败 %s: %v", file, err)
			}

			// 读取文件内容
			data, err := os.ReadFile(absFile)
			if err != nil {
				return nil, fmt.Errorf("读取配置文件失败 %s: %v", file, err)
			}

			// 解析 YAML 到目标结构体
			if err := yaml.Unmarshal(data, dest); err != nil {
				return nil, fmt.Errorf("解析配置文件失败 %s: %v", file, err)
			}

			xlog.Info("load config", xlog.String("file", file))
			loadedFiles = append(loadedFiles, absFile)
		}
	}

	// 处理环境变量替换
	if err := env.Parse(dest); err != nil {
		return nil, fmt.Errorf("处理环境变量失败: %v", err)
	}

	return uniqueStrings(loadedFiles), nil
}

// InitConf 初始化配置
func InitConf(dest any) error {
	if !xutil.IsPtr(dest) {
		return fmt.Errorf("配置目标必须是结构体指针类型")
	}
	confFiles := getConfFile()
	loadedFiles, err := fileConf(dest, confFiles...)
	if err != nil {
		return err
	}

	if !xenv.IsDev() {
		closeConfWatcher()
		return nil
	}

	xlog.Info("init config", xlog.String("files", strings.Join(confFiles, ",")))
	// 创建文件监听器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("创建监听器失败: %v", err)
	}

	for _, file := range loadedFiles {
		if err := watcher.Add(file); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			_ = watcher.Close()
			return fmt.Errorf("添加文件监听失败 %s: %v", file, err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	xutil.Go(ctx, func() {
		defer close(done)
		defer watcher.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					xlog.Error("监听器退出")
					return
				}
				// 如果是写入或创建事件，重新加载配置
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					if _, err := fileConf(dest, confFiles...); err != nil {
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
	})
	replaceConfWatcher(&confWatcherState{cancel: cancel, done: done})

	return nil
}

func CloseConfWatcher() {
	closeConfWatcher()
}

func closeConfWatcher() {
	replaceConfWatcher(nil)
}

func replaceConfWatcher(next *confWatcherState) {
	confWatcherMu.Lock()
	prev := currentConfWatch
	currentConfWatch = next
	confWatcherMu.Unlock()

	if prev != nil {
		prev.cancel()
		<-prev.done
	}
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
