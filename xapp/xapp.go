package xapp

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/daodao97/xgo/xlog"
)

type Startup func() error

func StartUpWarp(err error) Startup {
	return func() error {
		return err
	}
}

type BeforeStart func()

type Server interface {
	Start() error
	Stop()
}

type App struct {
	startups    []Startup
	servers     []NewServer
	beforeStart []BeforeStart
}

func NewApp() *App {
	return &App{}
}

func (a *App) AddStartup(startup ...Startup) *App {
	a.startups = append(a.startups, startup...)
	return a
}

func (a *App) AddServer(server ...NewServer) *App {
	a.servers = append(a.servers, server...)
	return a
}

func (a *App) AddBeforeStart(fn ...BeforeStart) *App {
	a.beforeStart = append(a.beforeStart, fn...)
	return a
}

func (a *App) Run() error {
	// 执行所有 Startup 函数
	for _, startup := range a.startups {
		if err := startup(); err != nil {
			return fmt.Errorf("startup error: %w", err)
		}
	}

	// 执行所有 BeforeStart 函数
	for _, fn := range a.beforeStart {
		fn()
	}

	// 启动所有 Server
	errChan := make(chan error, len(a.servers))
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var servers []Server

	for _, server := range a.servers {
		wg.Add(1)
		go func(s NewServer) {
			defer wg.Done()
			_s := s()
			servers = append(servers, _s)
			if err := _s.Start(); err != nil {
				errChan <- err
				cancel()
			}
		}(server)
	}

	// 处理信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待错误或信号
	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	case sig := <-sigChan:
		xlog.Debug("received signal", xlog.Any("signal", sig))
	case <-ctx.Done():
		xlog.Warn("Context cancelled")
	}

	// 优雅关闭
	xlog.Debug("Shutting down servers...", xlog.Any("num", len(servers)))
	for _, server := range servers {
		server.Stop()
	}

	wg.Wait()
	xlog.Debug("All servers stopped")
	return nil
}