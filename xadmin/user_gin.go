package xadmin

import "github.com/gin-gonic/gin"

func GinUserRoute(r *gin.RouterGroup) {
	r.POST("/user/login", httpHandlerToGin(loginHandler))
	r.POST("/user/email/code", httpHandlerToGin(emailCodeHandler))
	r.GET("/user/email/code/status", httpHandlerToGin(emailCodeStatusHandler))
	r.GET("/user/info", httpHandlerToGin(infoHandler))
	r.GET("/user/routes", httpHandlerToGin(routesHandler))
	r.GET("/user/logout", httpHandlerToGin(logoutHandler))
	r.GET("/user/form_mutex", httpHandlerToGin(formMutexHandler))
}
