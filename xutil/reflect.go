package xutil

import "reflect"

func IsPtr(v any) bool {
	ptrRef := reflect.ValueOf(v)
	if ptrRef.Kind() != reflect.Ptr {
		return false
	}
	ref := ptrRef.Elem()
	return ref.Kind() == reflect.Struct
}
