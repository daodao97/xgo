package xhttp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudflare/tableflip"

	"github.com/daodao97/xgo/xlog"
)

func StartServer(handler http.Handler, addr string, shutdownTimeout time.Duration) {
	upg, err := tableflip.New(tableflip.Options{})
	if err != nil {
		panic(err)
	}
	defer upg.Stop()

	// Do an upgrade on SIGHUP
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP)
		for range sig {
			err := upg.Upgrade()
			if err != nil {
				log.Println("Upgrade failed:", err)
			}
		}
	}()

	// Listen must be called before Ready
	ln, err := upg.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Can't listen:", err)
	}

	server := http.Server{
		Handler: handler,
		// Set timeouts, etc.
	}

	go func() {
		err := server.Serve(ln)
		if !errors.Is(err, http.ErrServerClosed) {
			log.Println("HTTP server:", err)
		}
	}()

	xlog.Info(fmt.Sprintf("serving on %s", ln.Addr()))
	if err := upg.Ready(); err != nil {
		panic(err)
	}
	<-upg.Exit()
	xlog.Info("shutting down...")

	shutdownServer(shutdownTimeout, server.Shutdown, os.Exit)
}

func shutdownServer(shutdownTimeout time.Duration, shutdownFn func(context.Context) error, exitFn func(int)) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := shutdownFn(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println("Graceful shutdown timed out")
			exitFn(1)
			return
		}
		log.Println("Graceful shutdown failed:", err)
	}
}
