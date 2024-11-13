package xapp

type Empty struct{}

type Page struct {
	Page int `form:"page" query:"page" default:"1" validate:"required,min=1"`
	Size int `form:"size" query:"size" default:"10" validate:"required,min=10"`
}

type PageResult[T any] struct {
	Page
	Total int `json:"total"`
	Items []T `json:"items"`
}
