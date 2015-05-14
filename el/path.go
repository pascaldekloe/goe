package el

import (
	"path"
	"reflect"
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
	o := reflect.ValueOf(root)
	for _, segment := range path {
		if segment == "" {
			break
		}

		k := o.Kind()
		for k == reflect.Ptr || k == reflect.Interface {
			o = o.Elem()
			k = o.Kind()
		}

		switch k {
		case reflect.Struct:
			o = o.FieldByName(segment)
		default:
			return
		}
	}

	k := o.Kind()
	for k == reflect.Ptr || k == reflect.Interface {
		o = o.Elem()
		k = o.Kind()
	}

	result = append(result, o)
	return
}
