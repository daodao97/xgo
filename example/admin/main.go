package main

import (
	"fmt"
	"net/http"

	"github.com/daodao97/xgo/cache"
	"github.com/daodao97/xgo/xapp"
	"github.com/daodao97/xgo/xdb"
	"github.com/daodao97/xgo/xlog"
	"github.com/daodao97/xgo/xredis"
	"github.com/gin-gonic/gin"

	"egg/conf"
	"egg/inerval/admin"
	"egg/inerval/api"
	"egg/inerval/auth"
	"egg/inerval/db"
)

var Version string

func main() {
	app := xapp.NewApp().
		AddStartup(
			conf.InitConf,
			func() error {
				return xredis.Init(conf.Get().Redis)
			},
			func() error {
				return db.InitDB(conf.Get().Database)
			},
		).
		AddBeforeStart(func() {
			xdb.SetCache(cache.NewRedis(xredis.Get(), cache.WithPrefix("egg")))
		}).
		AfterStarted(func() {
			xlog.Debug("version", xlog.String("version", Version))
		}).
		AddServer(xapp.NewGinHttpServer(xapp.Args.Bind, h))

	if err := app.Run(); err != nil {
		fmt.Printf("Application error: %v\n", err)
	}
}

func h() *gin.Engine {
	e := xapp.NewGin()
	defer openapi(e)

	e.Any("/ping", func(c *gin.Context) {
		xlog.DebugCtx(c, "ok")
		c.String(http.StatusOK, "pong")
	})
	e.POST("/login", xapp.RegisterAPI(api.Login))

	e.Group("/").Use(auth.AuthMiddleware())

	// other api

	admin.Route(e)
	return e
}

func openapi(e *gin.Engine) {
	_, err := xapp.GenerateOpenAPIDoc(e,
		xapp.WithBearerAuth(),
		xapp.WithInfo("Egg", "1.0.0"),
		xapp.WithServer("http://127.0.0.1:3003", "Local"),
	)
	if err != nil {
		xlog.Error("Failed to generate OpenAPI doc", xlog.Err(err))
	}
}
