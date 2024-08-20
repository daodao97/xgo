package xadmin

import (
	_ "embed"
	"net/http"

	"github.com/golang-jwt/jwt/v5"

	"github.com/daodao97/xgo/xdb"
	"github.com/daodao97/xgo/xhttp"
	"github.com/daodao97/xgo/xjwt"
	"github.com/daodao97/xgo/xlog"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"muzzammil.xyz/jsonc"
)

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

func UserRoute(r *mux.Router) {
	apiRouter := r.PathPrefix("/user").Subrouter()

	apiRouter.HandleFunc("/login", loginHandler).Methods("POST")
	apiRouter.HandleFunc("/info", infoHandler).Methods("GET")
	apiRouter.HandleFunc("/routes", routesHandler).Methods("GET")
	apiRouter.HandleFunc("/logout", logoutHandler).Methods("GET")
	apiRouter.HandleFunc("/form_mutex", formMutexHandler).Methods("GET")
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

	row := xdb.New("operator").SelectOne(xdb.WhereEq("username", user.Username))
	if row.Err != nil {
		xhttp.ResponseJson(w, Map{
			"code":    4001,
			"message": "用户名或密码错误",
		})
		return
	}

	//hash, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	//fmt.Println(string(hash), user.Password, row.GetString("password"))

	err = bcrypt.CompareHashAndPassword([]byte(row.GetString("password")), []byte(user.Password))
	if err != nil {
		xlog.Error("bcrypt.CompareHashAndPassword", xlog.Err(err))
		xhttp.ResponseJson(w, Map{
			"code":    4002,
			"message": "用户名或密码错误",
		})
		return
	}

	token, err := xjwt.GenHMacToken(jwt.MapClaims{
		"username": row.GetString("username"),
		"user_id":  row.GetInt("id"),
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
			"env":      "prod",
			"website":  website,
		},
	})

}

func routesHandler(w http.ResponseWriter, r *http.Request) {
	var data any
	err := jsonc.Unmarshal([]byte(routes), &data)
	if err != nil {
		xlog.Error("xadmin route unmarshal error", xlog.Err(err))
	}

	xhttp.ResponseJson(w, Map{
		"code": 0,
		"data": data,
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
