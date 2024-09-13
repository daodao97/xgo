package main

import (
	"net/http"
	"net/url"

	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/daodao97/xgo/xapp"
	"github.com/daodao97/xgo/xlog"
	"github.com/daodao97/xgo/xproxy"
	"github.com/gin-gonic/gin"

	"fmt"

	"github.com/bogdanfinn/tls-client/profiles"
)

var Vars struct {
	Origin    string `long:"origin" env:"ORIGIN" description:"Origin"`
	HttpProxy string `long:"http-proxy" env:"HTTP_PROXY" description:"Http proxy"`
}

func Proxy(client tls_client.HttpClient) http.Handler {
	endpointRouter := xapp.NewGin()

	endpointRouter.NoRoute(func(c *gin.Context) {
		resp, err := xproxy.Request(client, Vars.Origin, c)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.DataFromReader(resp.StatusCode, -1, resp.Header.Get("Content-Type"), resp.Body, nil)
	})

	return endpointRouter
}

func main() {
	xapp.ParserFlags(&Vars)

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(600),
		tls_client.WithClientProfile(profiles.Chrome_120),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithRandomTLSExtensionOrder(),
	}

	if Vars.HttpProxy != "" {
		_, err := url.Parse(Vars.HttpProxy)
		if err != nil {
			panic(fmt.Errorf("invalid proxy url: %v", err))
		}
		options = append(options, tls_client.WithProxyUrl(Vars.HttpProxy))
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		panic(fmt.Errorf("create http client error: %v", err))
	}

	xlog.Debug("xproxy", xlog.Any("origin", Vars.Origin), xlog.Any("proxy", Vars.HttpProxy))

	app := xapp.
		NewApp().
		AddServer(xapp.NewHttp(xapp.Args.Bind, func() http.Handler { return Proxy(client) }))

	if err := app.Run(); err != nil {
		fmt.Printf("Application error: %v\n", err)
	}
}
