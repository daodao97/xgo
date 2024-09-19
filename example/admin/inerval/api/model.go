package api

type LoginReq struct {
	Phone string `json:"phone" binding:"required,phone" form:"phone"`
	Code  string `json:"code" binding:"required" form:"code"`
}

type LoginResp struct {
	Token string `json:"jwt"`
}
