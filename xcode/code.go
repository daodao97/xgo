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
	return json.Marshal(c)
}

func (c *Code) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, c)
}
