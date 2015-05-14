package el

import (
	"path"
	"reflect"
	"strconv"
	"strings"
)

func eval(expr string, root interface{}) (result []reflect.Value) {
	expr = path.Clean(expr)
	if !path.IsAbs(expr) {
		return // TODO
	}

	return resolve(strings.Split(expr, "/")[1:], root)
}

func resolve(path []string, root interface{}) (result []reflect.Value) {
	v := reflect.ValueOf(root)
	for _, segment := range path {
		if segment == "" {
			break
		}

		var field, key string
		if i := strings.IndexByte(segment, '['); i < 0 {
			field = segment
		} else {
			last := len(segment) - 1
			if segment[last] != ']' {
				return
			}
			key = segment[i+1 : last]
			field = segment[:i]
		}

		if field != "" {
			v = follow(v)
			switch v.Kind() {
			case reflect.Struct:
				v = v.FieldByName(field)
			default:
				return
			}
		}

		if key != "" {
			v = follow(v)
			switch v.Kind() {
			case reflect.Slice, reflect.Array, reflect.String:
				i, err := strconv.Atoi(key)
				if err != nil || i < 0 || i >= v.Len() {
					return
				}
				v = v.Index(i)
			default:
				return
			}
		}
	}

	v = follow(v)
	result = append(result, v)
	return
}

func follow(v reflect.Value) reflect.Value {
	k := v.Kind()
	for k == reflect.Ptr || k == reflect.Interface {
		v = v.Elem()
		k = v.Kind()
	}
	return v
}
