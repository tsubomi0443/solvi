package logutils

import (
	"fmt"
	"log/slog"
	"reflect"
)

func LogAny(key string, v any) slog.Attr {
	if hasUnsupportedType(reflect.TypeOf(v)) {
		return slog.String(key, fmt.Sprintf("<%T>", v))
	}

	return slog.Any(key, v)
}

func hasUnsupportedType(t reflect.Type) bool {
	if t == nil {
		return false
	}

	for t.Kind() == reflect.Pointer {
		t = t.Elem()
		if t == nil {
			return false
		}
	}

	switch t.Kind() {
	case reflect.Func,
		reflect.Chan,
		reflect.UnsafePointer:
		return true

	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if hasUnsupportedType(t.Field(i).Type) {
				return true
			}
		}

	case reflect.Slice,
		reflect.Array:
		return hasUnsupportedType(t.Elem())

	case reflect.Map:
		return hasUnsupportedType(t.Key()) ||
			hasUnsupportedType(t.Elem())
	}

	return false
}
