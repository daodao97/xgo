package util

import (
	"github.com/daodao97/xgo/internal/jsonc"
	"github.com/pkg/errors"
)

func JsonStrRemoveComments(str string) (string, error) {
	jc := jsonc.ToJSON([]byte(str))
	if jsonc.Valid(jc) {
		return string(jc), nil
	}
	return "", errors.New("Invalid JSON")
}
