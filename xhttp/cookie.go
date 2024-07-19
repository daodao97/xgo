package xhttp

import (
	"net/http"
	"time"
)

type MyCookie struct {
	Key string
	Val string
}

func Cookie(r *http.Request, key string) string {
	cookie, err := r.Cookie(key)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func SetCookie(res http.ResponseWriter, cookies ...MyCookie) {
	for _, v := range cookies {
		http.SetCookie(res, &http.Cookie{
			Name:  v.Key,
			Value: v.Val,
			Path:  "/",
		})
	}
}

func DelCookie(res http.ResponseWriter, key ...string) {
	for _, v := range key {
		http.SetCookie(res, &http.Cookie{
			Name:    v,
			Value:   "",
			Path:    "/",
			Expires: time.Unix(0, 0), // 设置为1970年1月1日
			MaxAge:  -1,
		})

	}
}

// ExpireCookies 设置所有 cookie 为过期状态
func ExpireCookies(w http.ResponseWriter, r *http.Request) {
	// 获取请求中的所有 cookie
	cookies := r.Cookies()

	// 遍历 cookies，将它们的过期时间设置为过去的时间
	for _, cookie := range cookies {
		// 创建一个同名的新 cookie
		expiredCookie := &http.Cookie{
			Name:    cookie.Name,
			Value:   "",
			Expires: time.Unix(0, 0), // 将时间设置为 Unix 时间戳的起始点
			MaxAge:  -1,              // 或使用 MaxAge 设置为 -1
		}
		// 添加该 cookie 到响应中，覆盖客户端的旧 cookie
		http.SetCookie(w, expiredCookie)
	}
}

func RemoveReqCookie(req *http.Request, cookieName string) {
	// Get the current cookies
	cookies := req.Cookies()

	// Create a slice to hold the new set of cookies
	var newCookies []*http.Cookie

	// Loop through the current cookies
	for _, cookie := range cookies {
		// Add all cookies except the one to be removed
		if cookie.Name != cookieName {
			newCookies = append(newCookies, cookie)
		}
	}

	// Clear the existing Cookie header
	req.Header.Del("Cookie")

	// Add the new set of cookies to the request
	for _, cookie := range newCookies {
		req.AddCookie(cookie)
	}
}

func AddReqCookie(req *http.Request, cookieName string, value string) {
	req.AddCookie(&http.Cookie{Name: cookieName, Value: value})
}
