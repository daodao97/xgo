package xapp

type Empty struct{}

type Page struct {
	Page int `json:"page" query:"page" default:"1"`
	Size int `json:"size" query:"size" default:"10"`
}

type PageResult[T any] struct {
	Page
	Total int `json:"total"`
	Items []T `json:"items"`
}
