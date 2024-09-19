package admin

import (
	"embed"
	_ "embed"

	"github.com/daodao97/xgo/xadmin"
	"github.com/gin-gonic/gin"

	"egg/conf"
)

//go:embed route.jsonc
var routes string

//go:embed schema
var schema embed.FS

func Route(e *gin.Engine) *gin.RouterGroup {
	xadmin.SetRoutes(routes)
	xadmin.InitSchema(schema)
	xadmin.SetAdminPath(conf.Get().AdminPath)
	xadmin.SetJwt(&xadmin.JwtConf{
		Secret:      "abc",
		TokenExpire: 3600,
	})
	xadmin.SetWebSite(map[string]any{
		"title":         "xadmin",
		"logo":          "https://dow.chatbee.cc/chatgpt.jpeg",
		"defaultAvatar": "https://dow.chatbee.cc/chatgpt.jpeg",
		"formMutex":     false,
		"captcha":       false,
	})
	RegHook()
	admin := xadmin.GinRoute(e)

	return admin
}
