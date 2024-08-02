package xtype

import (
	"math"
	"strconv"
	"strings"
)

const EPSILON float64 = 1e-9
const FalseStrings = "no,false,off,0,"

// Bool try to get bool value of given object.
// number: 0 => false, otherwise => true
// string: ("", "false", "off", "no", "0") => false (case insensitive), otherwise => true
// nil: false
// otherwise: true
func Bool(obj any) bool {
	if obj == nil {
		return false
	}

	switch v := obj.(type) {
	case bool:
		return v
	case float32:
		return math.Abs(float64(v)) > EPSILON
	case float64:
		return math.Abs(v) > EPSILON
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return v != 0
	case []byte:
		s := strings.ToLower(string(v))
		return s != "" && !strings.Contains(FalseStrings, s+",")
	case string:
		s := strings.ToLower(v)
		return s != "" && !strings.Contains(FalseStrings, s+",")
	default:
		return true
	}
}

func Int64ToBinaryIndex(i int64) []int64 {
	var arr []int64
	str := strconv.FormatInt(i, 2)
	for i := range str {
		if str[i:i+1] == "1" {
			arr = append(arr, int64(i))
		}
	}
	return arr
}

func BinaryIndexToInt64(b []int64) int64 {
	arrStr := []string{"0", "0", "0", "0", "0", "0"}
	for _, j := range b {
		for i := 0; i < 6; i++ {
			if int64(i) == j {
				arrStr[i] = "1"
			}
		}
	}
	str := strings.Join(arrStr, "")
	i, _ := strconv.ParseInt(str, 2, 10)
	return i
}

func IsString(v any) bool {
	switch v.(type) {
	case string:
		return true
	}
	return false
}
