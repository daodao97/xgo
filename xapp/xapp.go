package xapp

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/daodao97/xgo/xlog"
	"github.com/jessevdk/go-flags"
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

var Args struct {
	Bind          string `long:"bind" description:"Bind address" default:"127.0.0.1:4001" env:"BIND"`
	EnableOpenAPI bool   `long:"enable-openapi" description:"Enable OpenAPI" env:"ENABLE_OPENAPI"`
	AppEnv        string `long:"app-env" description:"App environment" env:"APP_ENV" default:"dev"`
}

func init() {
	ParserFlags(&Args)
}

type App struct {
	startups     []Startup
	servers      []NewServer
	beforeStart  []BeforeStart
	afterStarted func()
}

func NewApp() *App {
	xlog.Debug("app args", xlog.Any("args", fmt.Sprintf("%+v", Args)))
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

func (a *App) AfterStarted(fn func()) *App {
	a.afterStarted = fn
	return a
}

func (a *App) Run() error {
	// 执行所有 Startup 函数
	for _, startup := range a.startups {
		if err := startup(); err != nil {
			xlog.Error("startup error", xlog.Err(err))
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

	if a.afterStarted != nil {
		a.afterStarted()
	}

	// 处理信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var signalCount int
	shutdownChan := make(chan struct{})

	// 启动一个 goroutine 处理信号
	go func() {
		for sig := range sigChan {
			signalCount++
			if signalCount >= 2 {
				xlog.Warn("received multiple signals, force exit",
					xlog.Any("signal", sig),
					xlog.Int("count", signalCount))
				os.Exit(1) // 强制退出
			}
			xlog.Debug("received signal",
				xlog.Any("signal", sig),
				xlog.Int("count", signalCount))
			if signalCount == 1 {
				close(shutdownChan)
			}
		}
	}()

	// 等待错误或信号
	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	case <-shutdownChan:
		xlog.Debug("Starting graceful shutdown...")
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

func ParserFlags(dest any) {
	parser := flags.NewParser(dest, flags.IgnoreUnknown)
	_, err := parser.Parse()
	if err != nil {
		fmt.Println("parse flags error")
		os.Exit(1)
	}
}
