package xadmin

import (
	_ "embed"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/daodao97/xgo/xdb"
	"github.com/daodao97/xgo/xhttp"
	"github.com/daodao97/xgo/xjwt"
	"github.com/daodao97/xgo/xlog"

	"github.com/tidwall/gjson"
	"golang.org/x/crypto/bcrypt"
	"muzzammil.xyz/jsonc"
)

var operatorTable = "operator"

func SetOperatorTable(table string) {
	operatorTable = table
}

var routes string

func SetRoutes(r string) {
	routes = r
}

type JwtConf struct {
	Secret      string
	TokenExpire int64
}

var _jwtConf *JwtConf

func SetJwt(c *JwtConf) {
	_jwtConf = c
}

var website = map[string]any{
	"title": "X-Admin",
}

func SetWebSite(data map[string]any) {
	website = data
}

type User struct {
	Username string
	Password string
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	user, err := xhttp.DecodeBody[User](r)
	if err != nil {
		xhttp.ResponseJson(w, Map{
			"code": 400,
			"msg":  "Invalid request",
		})
		return
	}

	row, err := xdb.New(operatorTable).Single(xdb.WhereEq("username", user.Username))
	if err != nil {
		xhttp.ResponseJson(w, Map{
			"code":    4001,
			"message": "用户名或密码错误",
		})
		return
	}

	if !row.GetBool("status") {
		xhttp.ResponseJson(w, Map{
			"code":    4003,
			"message": "用户已禁用",
		})
		return
	}

	if !PasswordVerify(user.Password, row.GetString("password")) {
		xhttp.ResponseJson(w, Map{
			"code":    4002,
			"message": "用户名或密码错误",
		})
		return
	}

	token, err := xjwt.GenHMacToken(jwt.MapClaims{
		"username": row.GetString("username"),
		"user_id":  row.GetInt("id"),
		"role":     row.GetString("role"),
	}, _jwtConf.Secret)
	if err != nil {
		xhttp.ResponseJson(w, Map{
			"code": 500,
			"msg":  "Generate token failed",
		})
		return
	}

	xhttp.ResponseJson(w, Map{
		"code": 0,
		"data": map[string]string{
			"name":  user.Username,
			"token": token,
		},
	})
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("x-token")
	if token == "" {
		xhttp.ResponseJson(w, Map{
			"code": 401,
			"msg":  "Unauthorized",
		})
		return
	}

	payload, err := xjwt.VerifyHMacToken(token, _jwtConf.Secret)
	if err != nil {
		xhttp.ResponseJson(w, Map{
			"code": 401,
			"msg":  "Unauthorized" + err.Error(),
		})
		return
	}

	xhttp.ResponseJson(w, Map{
		"code": 0,
		"data": Map{
			"id":       1,
			"name":     payload["username"].(string),
			"resource": nil,
			"env":      os.Getenv("APP_ENV"),
			"website":  website,
		},
	})

}

func routesHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("x-token")
	if token == "" {
		xhttp.ResponseJson(w, Map{
			"code": 401,
			"msg":  "Unauthorized",
		})
		return
	}

	payload, err := xjwt.VerifyHMacToken(token, _jwtConf.Secret)
	if err != nil {
		xhttp.ResponseJson(w, Map{
			"code": 401,
			"msg":  "Unauthorized",
		})
		return
	}

	userID, ok := payload["user_id"].(float64)
	if !ok {
		xhttp.ResponseJson(w, Map{
			"code": 401,
			"msg":  "Unauthorized",
		})
		return
	}

	row, err := xdb.New(operatorTable).Single(xdb.WhereEq("id", int(userID)))
	if err != nil {
		xhttp.ResponseJson(w, Map{
			"code": 401,
			"msg":  "Unauthorized",
		})
		return
	}

	roles := parseRoleSet(row.GetString("role"))

	jsonBytes := jsonc.ToJSON([]byte(routes))
	if !gjson.ValidBytes(jsonBytes) {
		xlog.Error("xadmin routes json invalid")
		xhttp.ResponseJson(w, Map{
			"code": 0,
			"data": []any{},
		})
		return
	}

	parsed := gjson.ParseBytes(jsonBytes)

	var filtered any
	if _, ok := roles["root"]; ok {
		filtered = parsed.Value()
	} else {
		filtered = filterRoutesResult(parsed, roles)
		if filtered == nil {
			filtered = []any{}
		}
	}

	xhttp.ResponseJson(w, Map{
		"code": 0,
		"data": filtered,
	})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	xhttp.ExpireCookies(w, r)
	xhttp.ResponseJson(w, Map{
		"code": 0,
	})
}

func formMutexHandler(w http.ResponseWriter, r *http.Request) {
	xhttp.ResponseJson(w, Map{
		"code": 0,
	})
}

func PasswordHash(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func PasswordVerify(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func parseRoleSet(role string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, item := range strings.Split(role, ",") {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		result[trimmed] = struct{}{}
	}
	return result
}

func filterRoutesResult(result gjson.Result, roles map[string]struct{}) any {
	if !result.Exists() {
		return nil
	}

	if result.IsArray() {
		items := result.Array()
		filtered := make([]any, 0, len(items))
		for _, item := range items {
			switch {
			case item.IsObject():
				if obj, ok := filterRouteObject(item, roles); ok {
					filtered = append(filtered, obj)
				}
			case item.IsArray():
				nested := filterRoutesResult(item, roles)
				if nested != nil {
					filtered = append(filtered, nested)
				}
			default:
				filtered = append(filtered, item.Value())
			}
		}
		return filtered
	}

	if result.IsObject() {
		if obj, ok := filterRouteObject(result, roles); ok {
			return obj
		}
		return nil
	}

	return result.Value()
}

func filterRouteObject(route gjson.Result, roles map[string]struct{}) (map[string]any, bool) {
	roleValue := route.Get("role")
	if roleValue.Exists() && !routeAllowedResult(roleValue, roles) {
		return nil, false
	}

	data := route.Map()
	filtered := make(map[string]any, len(data))
	for key, value := range data {
		if key == "routes" {
			nested := filterRoutesResult(value, roles)
			if nested == nil && value.Exists() {
				filtered[key] = []any{}
				continue
			}
			if nested != nil {
				filtered[key] = nested
			}
			continue
		}
		filtered[key] = value.Value()
	}

	return filtered, true
}

func routeAllowedResult(roleValue gjson.Result, roles map[string]struct{}) bool {
	if len(roles) == 0 {
		return false
	}

	if roleValue.Type == gjson.String {
		for _, item := range strings.Split(roleValue.Str, ",") {
			trimmed := strings.TrimSpace(item)
			if trimmed == "" {
				continue
			}
			if _, ok := roles[trimmed]; ok {
				return true
			}
		}
		return false
	}

	if roleValue.IsArray() {
		for _, item := range roleValue.Array() {
			if item.Type != gjson.String {
				continue
			}
			trimmed := strings.TrimSpace(item.Str)
			if trimmed == "" {
				continue
			}
			if _, ok := roles[trimmed]; ok {
				return true
			}
		}
	}

	return false
}
