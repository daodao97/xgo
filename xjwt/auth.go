package xjwt

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/daodao97/xgo/xjson"
	"github.com/gin-gonic/gin"
)

type authContextKey struct{}

// gin middleware
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// remove Bearer prefix
		token = strings.TrimPrefix(token, "Bearer ")

		payload, err := VerifyHMacToken(token, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized token invalid"})
			c.Abort()
			return
		}

		_payload := xjson.New(payload)
		if _payload.Get("exp").Int64() < time.Now().Unix() {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized token expired"})
			c.Abort()
			return
		}

		// 将 payload 添加到请求上下文
		ctx := context.WithValue(c.Request.Context(), authContextKey{}, _payload)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func GetAuthFromContext(ctx context.Context) xjson.Json {
	if _ctx, ok := ctx.(*gin.Context); ok {
		return GetAuth(_ctx)
	}
	auth, _ := ctx.Value(authContextKey{}).(xjson.Json)
	return auth
}

func GetAuth(c *gin.Context) xjson.Json {
	return GetAuthFromContext(c.Request.Context())
}
