package xapp

type Empty struct{}

type Page struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

type PageResult[T any] struct {
	Page
	Total int `json:"total"`
	Items []T `json:"items"`
}
