package xcode

import "encoding/json"

type Code struct {
	Code     int    `json:"code"`
	HttpCode int    `json:"http_code"`
	Message  string `json:"message"`
	Type     string `json:"type"`
	Err      error  `json:"-"`
}

func (c *Code) Error() string {
	return c.Message
}

func (c *Code) MarshalJSON() ([]byte, error) {
	// 创建一个新的结构体来序列化,避免递归
	type Alias Code
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	})
}

func (c *Code) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, c)
}
