package xadmin

import (
	"database/sql"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/daodao97/xgo/xdb"
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
	api.GET("/:table_name/get/:id", GinGet)
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
		cookieToken, _ := c.Cookie("oms%3Atoken")
		if cookieToken != "" {
			token = cookieToken
		}
		if token == "" {
			c.JSON(http.StatusOK, gin.H{"code": 401, "message": "Unauthorized: token is required"})
			c.Abort()
			return
		}

		payload, err := xjwt.VerifyHMacToken(token, _jwtConf.Secret)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    401,
				"message": "Unauthorized: " + err.Error(),
			})
			c.Abort()
			return
		}
		userId := payload["user_id"]

		user, err := xdb.New(operatorTable).Single(xdb.WhereEq("id", userId))
		if err != nil || !user.GetBool("status") {
			c.JSON(http.StatusOK, gin.H{
				"code":    401,
				"message": "Unauthorized: user not found or user is disabled",
			})
			c.Abort()
			return
		}

		c.Set("user", payload)

		c.Next()
	}
}

type DragSortRequest struct {
	Ids string `json:"ids"`
}

type DragSortMode string

const (
	DragSortModeAsc  DragSortMode = "asc"
	DragSortModeDesc DragSortMode = "desc"
)

func GinDragSort(m xdb.Model, sortField string, sortMode DragSortMode) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DragSortRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
			return
		}

		ids := strings.Split(req.Ids, ",")

		// 开启事务进行批量更新
		err := m.Transaction(func(tx *sql.Tx, model xdb.Model) error {
			for index, id := range ids {
				var sortValue int
				switch sortMode {
				case DragSortModeAsc:
					// 升序模式: 1, 2, 3, ...
					sortValue = index + 1
				case DragSortModeDesc:
					// 降序模式: n, n-1, n-2, ...
					sortValue = len(ids) - index
				default:
					// 默认使用升序
					sortValue = index + 1
				}

				_, err := model.Tx(tx).Update(map[string]any{
					sortField: sortValue,
				}, xdb.WhereEq("id", id))

				if err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
	}
}
