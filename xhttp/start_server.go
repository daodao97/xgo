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
	upg, err := tableflip.New(tableflip.Options{
		PIDFile: "pid",
	})
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

	// Make sure to set a deadline on exiting the process
	// after upg.Exit() is closed. No new upgrades can be
	// performed if the parent doesn't exit.
	time.AfterFunc(shutdownTimeout, func() {
		log.Println("Graceful shutdown timed out")
		os.Exit(1)
	})

	// Wait for connections to drain.
	server.Shutdown(context.Background())
}
