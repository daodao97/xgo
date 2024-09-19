package api

import (
	"regexp"

	"github.com/daodao97/xgo/xdb"
	"github.com/daodao97/xgo/xjwt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"

	"egg/conf"
)

// 添加自定义验证函数
func validateChinesePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	match, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
	return match
}

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("phone", validateChinesePhone)
	}
}

// @summary 登录
// @description 登录
func Login(c *gin.Context, req LoginReq) (*LoginResp, error) {
	var _login = func(req LoginReq) (*LoginResp, error) {
		token, err := login(c, req.Phone)
		if err != nil {
			return nil, err
		}

		return &LoginResp{Token: token}, nil
	}

	// 校验验证码

	return _login(req)
}

func login(c *gin.Context, phone string) (string, error) {
	m := xdb.New("user").Ctx(c)

	user, _, err := m.InsertOrUpdate(xdb.Record{"phone": phone})
	if err != nil {
		return "", err
	}

	var payload jwt.MapClaims = jwt.MapClaims(user)

	token, err := xjwt.GenHMacToken(payload, conf.Get().JwtSecret)
	if err != nil {
		return "", err
	}

	return token, nil
}
