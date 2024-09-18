package xadmin

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/daodao97/xgo/xjwt"
	"github.com/gin-gonic/gin"
)

func GinRoute(r *gin.Engine) *gin.RouterGroup {
	_ui := fs.FS(defaultUI)
	if customUI != nil {
		_ui = customUI
	}

	contentStatic, err := fs.Sub(_ui, "ui")
	if err != nil {
		panic(err)
	}

	// 创建静态文件服务
	r.StaticFS(adminPath, http.FS(contentStatic))

	api := r.Group(fmt.Sprintf("%sapi", adminPath))
	api.Use(authMiddleware())

	api.GET("/schema/:table_name", GinPageSchema)
	api.POST("/:table_name/create", GinCreate)
	api.GET("/:table_name/list", GinList)
	api.GET("/:table_name/get/:id", GinRead)
	api.POST("/:table_name/update/:id", GinUpdate)
	api.DELETE("/:table_name/del", GinDelete)
	api.GET("/:table_name/options", GinOptions)

	GinUserRoute(api)

	return api
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == fmt.Sprintf("%sapi/user/login", adminPath) {
			c.Next()
			return
		}

		token := c.GetHeader("X-Token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401})
			c.Abort()
			return
		}

		_, err := xjwt.VerifyHMacToken(token, _jwtConf.Secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "Unauthorized: " + err.Error(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
