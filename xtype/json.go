package xtype

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"muzzammil.xyz/jsonc"
)

func init() {
	once := sync.Once{}
	once.Do(func() {
		extra.RegisterFuzzyDecoders()
	})
}

func Binding(from any, to any) error {
	switch from := from.(type) {
	case []byte:
		return jsoniter.Unmarshal(from, to)
	case string:
		if from == "" {
			return fmt.Errorf("the source data is empty string")
		}
		return jsoniter.UnmarshalFromString(from, to)
	case io.ReadCloser:
		body, err := io.ReadAll(from)
		if err != nil {
			return err
		}
		return Binding(body, to)
	default:
		tmp, err := jsoniter.Marshal(from)
		if err != nil {
			return err
		}
		err = jsoniter.Unmarshal(tmp, to)
		if err != nil {
			return err
		}
		return nil
	}
}

func ToString(v any) string {
	ref := reflect.ValueOf(v)
	switch ref.Kind() {
	case reflect.Map, reflect.Slice, reflect.Struct:
		str, _ := json.Marshal(v)
		return string(str)
	default:
		return cast.ToString(v)
	}
}

func JsonStrVarReplace(jsonStr string, macro map[string]any) string {
	if macro == nil {
		return jsonStr
	}
	var macroVarReg = regexp.MustCompile(`(?U){{(.*)}}`)

	all := macroVarReg.FindAllString(jsonStr, -1)
	var localMacroKey [][]string
	for _, v := range all {
		match := macroVarReg.FindStringSubmatch(v)
		key := strings.TrimSpace(match[1])
		if String(key).StartWith(".") {
			localMacroKey = append(localMacroKey, []string{match[0], key})
			continue
		}
		tokens := String(key).Split("|")
		data, ok := macro[tokens.First().Raw()]
		if !ok {
			continue
		}
		rep := match[0]
		if tokens.Has("raw") {
			rep = "\"" + rep + "\""
		}
		jsonStr = strings.ReplaceAll(jsonStr, rep, ToString(data))
	}

	for _, v := range localMacroKey {
		value := gjson.Get(jsonStr, String(v[1]).ReplaceFirst(".", "").Split("|").First().Raw())
		rep := v[0]
		if String(v[1]).Split("|").Has("raw") {
			rep = "\"" + rep + "\""
		}
		jsonStr = strings.ReplaceAll(jsonStr, rep, value.String())
	}

	return jsonStr
}

func JsonStrRemoveComments(str string) (string, error) {
	jc := jsonc.ToJSON([]byte(str))
	if jsonc.Valid(jc) {
		return string(jc), nil
	}
	return "", errors.New("Invalid JSON")
}
