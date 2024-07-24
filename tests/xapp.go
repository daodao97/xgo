package main

import (
	"fmt"
	"net/http"

	"github.com/daodao97/xgo/xapp"
)

// 使用示例
func main() {
	app := xapp.NewApp().
		AddStartup(func() error {
			fmt.Println("Initializing configuration...")
			return nil
		}).
		AddServer(
			xapp.NewHttp(":8900", func() http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintf(w, "Hello, World!")
				})
			}),
			xapp.NewHttp(":8901", func() http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintf(w, "Welcome to the API!")
				})
			}),
		)

	if err := app.Run(); err != nil {
		fmt.Printf("Application error: %v\n", err)
	}
}
