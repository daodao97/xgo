package xtype

import (
	"encoding/json"
	"regexp"
	"strings"
)

type String string

func (s String) TrimSpace() String {
	return String(strings.TrimSpace(string(s)))
}

func (s String) StartWith(str string) bool {
	return strings.HasPrefix(string(s), str)
}

func (s String) EndWith(str string) bool {
	return strings.HasSuffix(string(s), str)
}

func (s String) Contains(str string) bool {
	return strings.Contains(string(s), str)
}

func (s String) Binding(to any) error {
	return json.Unmarshal([]byte(s), to)
}

func (s String) DecodeMap() *MapStrAny {
	tmp := new(MapStrAny)
	_ = json.Unmarshal([]byte(s), tmp)
	return tmp
}

func (s String) DecodeMapE() (*MapStrAny, error) {
	tmp := new(MapStrAny)
	err := json.Unmarshal([]byte(s), tmp)
	return tmp, err
}

func (s String) DecodeArrMap() *ArrMap {
	tmp := new(ArrMap)
	_ = json.Unmarshal([]byte(s), tmp)
	return tmp
}

func (s String) DecodeArrMapE() (*ArrMap, error) {
	tmp := new(ArrMap)
	err := json.Unmarshal([]byte(s), tmp)
	return tmp, err
}

func (s String) DecodeInterfaceE() (any, error) {
	tmp := new(any)
	err := json.Unmarshal([]byte(s), tmp)
	return tmp, err
}

func (s String) Split(sep string) *ArrStr {
	tmp := ArrStr(strings.Split(string(s), sep)).Filter(func(index int, str string) bool {
		return str != ""
	})
	return &tmp
}

func (s String) RegexSplit(str string) ArrStr {
	reg := regexp.MustCompile(str)
	split := reg.Split(string(s), -1)
	if len(split) == 0 {
		return []string{str}
	}
	var set []string

	for i := range split {
		set = append(set, split[i])
	}

	return set
}

func (s String) ReplaceFirst(old, new string) String {
	return String(strings.Replace(s.Raw(), old, new, 1))
}

func (s String) ReplaceAll(old, new string) String {
	return String(strings.ReplaceAll(s.Raw(), old, new))
}

func (s String) Raw() string {
	return string(s)
}
