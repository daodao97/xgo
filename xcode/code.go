package xcode

import (
	"encoding/json"
	"fmt"
)

type Code struct {
	Code     int    `json:"code"`
	HttpCode int    `json:"http_code"`
	Message  string `json:"message"`
	Type     string `json:"type"`
	Err      error  `json:"-"`
}

func (c *Code) Error() string {
	return fmt.Sprintf("code: %d, http_code: %d, message: %s, type: %s, err: %v", c.Code, c.HttpCode, c.Message, c.Type, c.Err)
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

// func (c *Code) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(c)
// }

func (c *Code) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, c)
}
